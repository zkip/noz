package main

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strings"
	"time"

	"noz.zkip.cc/utils"
)

func insertHierarchyRecord(parentID string, targetPRI string, name string, order ...uint) (string, error) {
	db := utils.GetMySqlDB()
	hierarchyID := genHierarchyID(targetPRI)

	isLastOrder := len(order) == 0

	var _order uint = 0
	if !isLastOrder {
		_order = order[0]
	}

	var err error

	tx, err := db.Begin()
	if err != nil {
		return "", err
	}

	rollback := func(err error) {
		if tx != nil {
			if err := tx.Rollback(); err != nil {
				panic(err)
			}
		}
	}

	_, err = tx.Exec("insert into tHierarchy( ancestor, descendant, distance, targetPRI ) ( select ancestor, ?, distance + 1, ? from tHierarchy where descendant = ? )", hierarchyID, targetPRI, parentID)
	if utils.RunIfErr(err, rollback) {
		return "", err
	}

	_, err = tx.Exec("insert into tHierarchy( ancestor, descendant, distance, targetPRI ) values( ?, ?, 0, ? )", hierarchyID, hierarchyID, targetPRI)
	if utils.RunIfErr(err, rollback) {
		return "", err
	}

	/*
		insert data
	*/
	ancestorSizeQtr := "select * from (select size from tHierarchyData where hierarchyID = ?) tmp"
	// ensure order is in safe bound.(0-parent_size)
	orderQtr := fmt.Sprintf("GREATEST(0, LEAST((%s), ?))", ancestorSizeQtr)
	if isLastOrder {
		orderQtr = fmt.Sprintf("%s where ? = 0", ancestorSizeQtr)
	}
	dataQtr := fmt.Sprintf("insert into tHierarchyData( hierarchyID, targetPRI, name, `order` ) values( ?, ?, ?, ifnull((%s), 0) )", orderQtr)
	_, err = tx.Exec(dataQtr, hierarchyID, targetPRI, name, parentID, _order)
	if utils.RunIfErr(err, rollback) {
		return "", err
	}

	/*
		update parent size
	*/
	_, err = tx.Exec("update tHierarchyData set size = size + 1 where hierarchyID = ?", parentID)
	if utils.RunIfErr(err, rollback) {
		return "", nil
	}

	/*
		update siblings order
	*/
	if !isLastOrder {
		idQtr := "select descendant from tHierarchy where ancestor = ? and distance = 1"
		siblingsQtr := fmt.Sprintf("update tHierarchyData set `order` = `order` + 1 where `order` >= ? and hierarchyID != ? and hierarchyID in (%s)", idQtr)
		_, err = tx.Exec(siblingsQtr, _order, hierarchyID, parentID)
		if utils.RunIfErr(err, rollback) {
			return "", err
		}
	}

	if err = tx.Commit(); err != nil {
		return "", err
	}

	return hierarchyID, nil
}
func findChildren(hierarchyID string) ([]string, error) {
	db := utils.GetMySqlDB()
	var descendants = make([]string, 0)

	rows, err := db.Query("select name from tHierarchyData where hierarchyID in (select descendant from tHierarchy where ancestor = ? and distance = 1)", hierarchyID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var descendant string

		err := rows.Scan(&descendant)
		if err != nil {
			return nil, err
		}

		descendants = append(descendants, descendant)
	}

	return descendants, nil
}

func findPath(hierarchyID string) ([]string, error) {
	db := utils.GetMySqlDB()
	var ancestors = make([]string, 0)

	rows, err := db.Query("select name from tHierarchyData where hierarchyID in (select ancestor from tHierarchy where descendant = ? order by distance desc)", hierarchyID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var descendant string

		err := rows.Scan(&descendant)
		if err != nil {
			return nil, err
		}

		ancestors = append(ancestors, descendant)
	}

	return ancestors, err
}

func deleteHierarchyRecord(hierarchyID string) (*HierarchyRecordDeleteBack, error) {
	db := utils.GetMySqlDB()

	var err error

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	rollback := func(err error) {
		if tx != nil {
			if err := tx.Rollback(); err != nil {
				panic(err)
			}
		}
	}

	// prepare order, parent_size and parent_ID
	var order uint
	var size uint
	var parentID string
	rows, err := tx.Query("select `order`, size, hierarchyID from tHierarchyData where hierarchyID = ? or hierarchyID = (select ancestor from tHierarchy where distance = 1 and descendant = ?)", hierarchyID, hierarchyID)
	if utils.RunIfErr(err, rollback) {
		return nil, err
	}

	noRecord := true
	noParent := true

	for rows.Next() {
		var order_, size_ uint
		var hID string

		err := rows.Scan(&order_, &size_, &hID)
		if utils.RunIfErr(err, rollback) {
			return nil, err
		}

		if hID == hierarchyID {
			order = order_
		} else {
			parentID = hID
			size = size_

			noParent = false
		}

		noRecord = false
	}

	if noRecord {
		return nil, nil
	}

	isLastOrder := order == size-1

	/*
		update siblings order
	*/
	if !isLastOrder {
		idQtr := "select descendant from tHierarchy where ancestor = ? and distance = 1"
		siblingsQtr := fmt.Sprintf("update tHierarchyData set `order` = `order` - 1 where `order` > ? and hierarchyID != ? and hierarchyID in (%s)", idQtr)
		_, err := tx.Exec(siblingsQtr, order, hierarchyID, parentID)
		if utils.RunIfErr(err, rollback) {
			return nil, err
		}
	}

	/*
		update parent size
	*/
	if !noParent {
		_, err = tx.Exec("update tHierarchyData set size = size - 1 where hierarchyID = ?", parentID)
		if utils.RunIfErr(err, rollback) {
			return nil, err
		}
	}

	/*
		delete related data
	*/
	var hierarchyIDCondQtr = ""
	var descendantCondQtr = ""
	var seperater = ""
	// get posterity
	rows, err = tx.Query("select descendant from tHierarchy where ancestor = ?", hierarchyID)
	if utils.RunIfErr(err, rollback) {
		return nil, err
	}

	for rows.Next() {
		var descendant string
		err = rows.Scan(&descendant)
		if utils.RunIfErr(err, rollback) {
			return nil, err
		}
		hierarchyIDCondQtr += seperater + fmt.Sprintf(`hierarchyID = "%s"`, descendant)
		descendantCondQtr += seperater + fmt.Sprintf(`descendant = "%s"`, descendant)
		seperater = " or "
	}

	_, err = tx.Exec(fmt.Sprintf("delete from tHierarchy where %s", descendantCondQtr))
	if utils.RunIfErr(err, rollback) {
		return nil, err
	}
	_, err = tx.Exec(fmt.Sprintf("delete from tHierarchyData where %s", hierarchyIDCondQtr))
	if utils.RunIfErr(err, rollback) {
		return nil, err
	}

	return &HierarchyRecordDeleteBack{ParentID: parentID}, tx.Commit()
}
func renameHierarchyRecord(hierarchyID string, name string) error {
	db := utils.GetMySqlDB()
	_, err := db.Exec("update tHierarchyData set name = ? where hierarchyID = ?", name, hierarchyID)
	if err != nil {
		return err
	}
	return nil
}

/*
	rough steps:

	in hierarchy table
	1. delete past id
	2. create new id

	in hierarchy data table
	3. update new siblings order ( +1 on >= new target order )
	4. update new parent size +1
	5. update past siblings order ( -1 on > past target order )
	6. update past parent size -1
	7. update target to new order
	8. update target to new id
*/
func moevHierarchyRecord(hierarchyID string, destHID string, order uint) error {
	db := utils.GetMySqlDB()

	var err error
	var qtr string

	tx, err := db.Begin()
	utils.PanicIfErr(err)

	rollback := func(err error) {
		if tx != nil {
			utils.PanicIfErr(tx.Rollback())
		}
	}

	var pastParentID, newParentID string
	var pastOrder, newOrder uint
	var pastSiblingCount, newSiblingCount uint
	var c = 0

	newOrder = order
	newParentID = destHID

	// find pastParentRecord, newParentRecord and pastRecord
	qtr = `-- move before query
		select d.size, d.order, d.hierarchyID, t.descendant from tHierarchyData d
		join tHierarchy t
		on ( t.ancestor = d.hierarchyID and t.distance = 1 and (	-- query parent
			t.descendant = ?										-- pastParentRecord
		))
		or ( t.descendant = d.hierarchyID and t.distance = 0 and (	-- query self
				d.hierarchyID = ?									-- newParentRecord
			or	d.hierarchyID = ?									-- pastRecord
		))
	`
	rows, err := tx.Query(qtr, hierarchyID, destHID, hierarchyID)
	if utils.RunIfErr(err, rollback) {
		return err
	}

	for rows.Next() {
		var size, order uint
		var ID, PID string
		err := rows.Scan(&size, &order, &ID, &PID)
		if utils.RunIfErr(err, rollback) {
			return err
		}

		if PID != ID { // pastParentRecord
			pastSiblingCount = size
			pastParentID = ID
			fmt.Println("pastParentRecord: ", pastSiblingCount)

			c++
		} else if ID == destHID { // newParentRecord
			newSiblingCount = size
			fmt.Println("newParentRecord: ", newSiblingCount)

			c++
		} else if ID == hierarchyID { // pastRecord
			pastOrder = order
			fmt.Println("pastRecord: ", pastOrder)

			c++
		}

	}

	var isSafeUpdate = c == 3

	if utils.RunIfOK(!isSafeUpdate, rollback) {
		// Unsupported change top record position
		return &UnsafeMoveErr{}
	}

	if newOrder >= newSiblingCount {
		newOrder = newSiblingCount
	}

	isSameParent := pastParentID == newParentID

	/*
		rebuild relation. ref from https://www.percona.com/blog/2011/02/14/moving-subtrees-in-closure-table
	*/
	qtr = strings.Join([]string{
		"delete a from tHierarchy as a",
		"join tHierarchy as d on a.descendant = d.descendant",
		"left join tHierarchy as x",
		"on x.ancestor = d.ancestor and x.descendant = a.ancestor",
		"where d.ancestor = ? and x.ancestor is null",
	}, " ")
	_, err = tx.Exec(qtr, hierarchyID)
	if utils.RunIfErr(err, rollback) {
		return err
	}

	qtr = strings.Join([]string{
		"insert into tHierarchy (ancestor, descendant, distance, targetPRI)",
		"select supertree.ancestor, subtree.descendant,",
		"supertree.distance + subtree.distance + 1, subtree.targetPRI",
		"from tHierarchy AS supertree join tHierarchy AS subtree",
		"where subtree.ancestor = ?",
		"and supertree.descendant = ?",
	}, " ")
	_, err = tx.Exec(qtr, hierarchyID, destHID)
	if utils.RunIfErr(err, rollback) {
		return err
	}

	isNewLastOne := newSiblingCount-newOrder == 1
	isPastLastOne := pastSiblingCount-pastOrder == 1

	if isSameParent {
		// avoid dirty writes

		orders := []int{int(pastOrder), int(newOrder)}
		sort.Ints(orders)

		minOrder := orders[0]
		maxOrder := orders[1]

		delta := -1
		if pastOrder > newOrder {
			delta = 1
		}

		idQtr := "select descendant from tHierarchy where ancestor = ? and distance = 1"
		qtr = fmt.Sprintf("update tHierarchyData set `order` = `order` + ? where `order` >= ? and `order` <= ? and hierarchyID != ? and hierarchyID in (%s)", idQtr)
		_, err = tx.Exec(qtr, delta, minOrder, maxOrder, hierarchyID, destHID)
		if utils.RunIfErr(err, rollback) {
			return err
		}
	} else {

		// update new siblings order
		if !isNewLastOne {
			idQtr := "select descendant from tHierarchy where ancestor = ? and distance = 1"
			qtr = fmt.Sprintf("update tHierarchyData set `order` = `order` + 1 where `order` >= ? and hierarchyID != ? and hierarchyID in (%s)", idQtr)
			_, err = tx.Exec(qtr, order, hierarchyID, destHID)
			if utils.RunIfErr(err, rollback) {
				return err
			}
		}

		// update past siblings order
		if !isPastLastOne {
			idQtr := "select descendant from tHierarchy where ancestor = ? and distance = 1"
			qtr = fmt.Sprintf("update tHierarchyData set `order` = `order` - 1 where `order` > ? and hierarchyID != ? and hierarchyID in (%s)", idQtr)
			_, err = tx.Exec(qtr, pastOrder, hierarchyID, pastParentID)
			if utils.RunIfErr(err, rollback) {
				return err
			}
		}
	}

	// update order
	_, err = tx.Exec("update tHierarchyData set `order` = ? where hierarchyID = ?", newOrder, hierarchyID)
	if utils.RunIfErr(err, rollback) {
		return err
	}

	/*
		update sizes of affected parent
	*/
	if !isSameParent {
		_, err = tx.Exec("update tHierarchyData set size = size + 1 where hierarchyID = ?", newParentID)
		if utils.RunIfErr(err, rollback) {
			return err
		}

		_, err = tx.Exec("update tHierarchyData set size = size - 1 where hierarchyID = ?", pastParentID)
		if utils.RunIfErr(err, rollback) {
			return err
		}
	}

	return tx.Commit()
}

func genHierarchyID(userPRI string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(userPRI+time.Now().String())))
}
