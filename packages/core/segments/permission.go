package main

import (
	"database/sql"
	"fmt"

	"github.com/twinj/uuid"
	"noz.zkip.cc/utils"
)

const (
	permission_owner_read = 1 << iota
	permission_owner_write

	permission_group_read
	permission_group_write

	permission_others_read
	permission_others_write

	permission_all = 1<<iota - 1
)

const (
	visibility_private = permission_owner_read | permission_owner_write
	visibility_group   = visibility_private | permission_group_read
	visibility_public  = visibility_group | permission_others_read
)

// Path Resource Identifier, RESOURCE_TYPE/ID
type permission struct {
	takerPRI    string
	resourcePRI string

	which uint8 // bitmap, owner_r owner_w group_r group_w others_r others_w
}

type group struct {
	kind    uint8
	id      uint64
	members []string
}

func (perm permission) canIRead() bool {
	return perm.which&permission_owner_read != 0
}
func (perm permission) canIWrite() bool {
	return perm.which&permission_owner_write != 0
}
func (perm permission) canGroupRead() bool {
	return perm.which&permission_group_read != 0
}
func (perm permission) canGroupWrite() bool {
	return perm.which&permission_group_write != 0
}
func (perm permission) canOthersRead() bool {
	return perm.which&permission_others_read != 0
}
func (perm permission) canOthersWrite() bool {
	return perm.which&permission_others_write != 0
}
func (perm permission) both(which uint8) bool {
	return perm.which&which == which
}
func (perm permission) some(which uint8) bool {
	return perm.which&which != 0
}
func (perm permission) canRead() bool {
	return perm.some(permission_owner_read | permission_group_read | permission_others_read)
}
func (perm permission) canWrite() bool {
	return perm.some(permission_owner_write | permission_group_write | permission_others_write)
}

func getPermissionForGivens(resourcePRI string, takerPRIs ...string) ([]*permission, error) {
	perms := make([]*permission, 0)

	takerQueryCondStr := "isNull(takerPRI)"
	separator := " or "

	for range takerPRIs {
		takerQueryCondStr += separator + "takerPRI = ?"
	}

	takerQueryCondStr += " and"

	db := utils.GetMySqlDB()

	var args []interface{} = utils.ToInterfaceSlice(takerPRIs)
	args = append(args, resourcePRI)

	queryStr := fmt.Sprintf("select which, takerPRI from tPermissions where %s resourcePRI = ?", takerQueryCondStr)
	rows, err := db.Query(queryStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var eigenPerm *permission

	for rows.Next() {
		perm := &permission{resourcePRI: resourcePRI}
		var takerPRINStr sql.NullString
		err := rows.Scan(&perm.which, &takerPRINStr)
		if err != nil {
			return nil, err
		}

		perm.takerPRI = takerPRINStr.String

		if perm.takerPRI == "" {
			eigenPerm = perm
		} else {
			perms = append(perms, perm)
		}
	}

	if eigenPerm != nil {
		perms = append(perms, eigenPerm)
	}

	return perms, nil
}

func mergePermission(perms ...*permission) *permission {
	perm := &permission{}
	for _, pm := range perms {
		perm.resourcePRI = pm.resourcePRI
		perm.which |= pm.which
	}
	return perm
}

type PermissionDeniedErr struct{}

func (err *PermissionDeniedErr) Error() string {
	return "Permission denied."
}

func getPermission(resourcePRI string, takerPRI string) (*permission, error) {
	db := utils.GetMySqlDB()
	takerRPIs := []string{takerPRI}

	inGroups := false

	ownerPRI, err := getResourceOwner(resourcePRI)
	if err != nil {
		return nil, err
	}

	isOwnerTaker := ownerPRI == takerPRI

	if !isOwnerTaker {
		takerRPIs = []string{}
		rows, err := db.Query("select gid from tPermissionGroups where memberPRI = ?", takerPRI)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var groupPRI string
			err := rows.Scan(&groupPRI)
			if err != nil {
				return nil, err
			}
			takerRPIs = append(takerRPIs, groupPRI)
			inGroups = true
		}
	}

	perms, err := getPermissionForGivens(resourcePRI, takerRPIs...)
	if err != nil {
		return nil, err
	}

	if len(perms) == 0 {
		return &permission{resourcePRI: resourcePRI, takerPRI: takerPRI}, nil
	}

	eigenPerm := perms[len(perms)-1]
	perms = perms[:len(perms)-1]

	specifiedPerm := mergePermission(perms...)

	specifiedPerm.resourcePRI = resourcePRI
	specifiedPerm.takerPRI = takerPRI

	// extract group permssion
	var mask uint8 = permission_all ^ (permission_group_read | permission_group_write)
	specifiedWhich := specifiedPerm.which | mask

	perm := eigenPerm
	perm.which &= specifiedWhich

	isGroupTaker := takerPRI != "" && !isOwnerTaker && inGroups

	// get account-related permission
	var arMask uint8 = 0
	if isOwnerTaker {
		arMask = permission_owner_read | permission_owner_write
	} else if isGroupTaker {
		arMask = permission_group_read | permission_group_write
	} else {
		arMask = permission_others_read | permission_others_write
	}

	perm.which &= arMask

	return perm, nil
}

type ResourceNonExistErr struct {
	pri string
}

func (re *ResourceNonExistErr) Error() string {
	return fmt.Sprintf("Resource (%s) don't exist.", re.pri)
}

func getResourceOwner(resourcePRI string) (string, error) {
	db := utils.GetMySqlDB()

	ID := utils.ExtractRID(resourcePRI)

	row := db.QueryRow("select ownerPRI from tResources where rid = ?", ID)

	var ownerPRI string

	err := row.Scan(&ownerPRI)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", &ResourceNonExistErr{resourcePRI}
		}
		return "", err
	}

	return ownerPRI, nil
}

// be changed by owner only supported currently
func setPermission(resourcePRI string, takerPRI string, perm uint8) error {
	db := utils.GetMySqlDB()

	owner, err := getResourceOwner(resourcePRI)
	if err != nil {
		return err
	}

	isOwnerTaker := owner == takerPRI

	if !isOwnerTaker {
		return &PermissionDeniedErr{}
	}

	_, err = db.Exec("update tPermissions set which = ? where resourcePRI = ? and isNull(takerPRI)", perm, resourcePRI)
	if err != nil {
		return err
	}

	return nil
}

type permPayload struct {
	permID  uint64
	groupID string
	which   uint8
}

func setSpecifyPermission(resourcePRI string, takerPRI string, perm uint8) error {
	db := utils.GetMySqlDB()

	// extract group permssion
	var mask uint8 = permission_group_read | permission_group_write
	groupPermWhich := perm & mask

	// find valid group ID

	rows, err := db.Query("select takerPRI, id, which from tPermissions where resourcePRI = ? and !isNull(takerPRI) ", resourcePRI)
	if err != nil {
		return err
	}

	var potentialGroupIDPermPayloadM = map[string](*permPayload){}

	var validGroupID string

	for rows.Next() {
		var pp = &permPayload{}
		var groupPRI string
		err := rows.Scan(&groupPRI, &pp.permID, &pp.which)
		if err != nil {
			return err
		}

		pp.groupID = utils.ExtractGID(groupPRI)

		if pp.which&mask == groupPermWhich {
			validGroupID = pp.groupID
		} else {
			potentialGroupIDPermPayloadM[pp.groupID] = pp
		}

	}

	hasGreatGroup := validGroupID != ""
	hasPotentialPerm := len(potentialGroupIDPermPayloadM) > 0

	// clean for permissions table
	if hasPotentialPerm {
		gidCondStr := ""
		permIDCondStr := ""
		seperator := ""
		for gid, pp := range potentialGroupIDPermPayloadM {
			gidCondStr += seperator + "gid = \"" + gid + "\""
			permIDCondStr += seperator + "id = " + utils.ToString(pp.permID)
			seperator = " or "
		}

		groupQueryStr := fmt.Sprintf("select memberPRI, gid from tPermissionGroups where %s", gidCondStr)
		rows, err := db.Query(groupQueryStr)

		if err != nil {
			return err
		}

		absolveGIDM := map[string]bool{}
		for rows.Next() {
			var gid string
			var memberPRI string
			if err := rows.Scan(&memberPRI, &gid); err != nil {
				return err
			}

			if memberPRI != takerPRI {
				absolveGIDM[gid] = true
			}
		}

		permGIDCondStr := ""
		seperator = ""
		noOneAbsolved := true

		// subtract of potential gorups and absolved groups
		for gid := range potentialGroupIDPermPayloadM {
			_, isAbsolved := absolveGIDM[gid]
			groupPRI := utils.GenPRI(gid, utils.Resource_type_group)

			if !isAbsolved {
				permGIDCondStr += seperator + "takerPRI = \"" + groupPRI + "\""
				seperator = " or "

				noOneAbsolved = false
			}
		}
		if !noOneAbsolved {
			_, err := db.Exec(fmt.Sprintf("delete from tPermissions where %s", permGIDCondStr))
			if err != nil {
				return err
			}
		}
	}

	if !hasGreatGroup {
		tx, err := db.Begin()
		if err != nil {
			return err
		}

		rollback := func() {
			if tx != nil {
				if err := tx.Rollback(); err != nil {
					panic(err)
				}
			}
		}

		validGroupID = uuid.NewV4().String()

		_, err = tx.Exec("insert into tPermissionGroups(memberPRI, gid) values(?, ?)", takerPRI, validGroupID)
		if err != nil {
			rollback()
			return err
		}

		groupPRI := utils.GenPRI(validGroupID, utils.Resource_type_group)

		_, err = db.Exec("insert into tPermissions(which, takerPRI, resourcePRI) values(?, ?, ?)", groupPermWhich, groupPRI, resourcePRI)
		if err != nil {
			rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			rollback()
			return err
		}
	} else if err != nil {

		return err
	}

	if hasGreatGroup {

		existed := true

		row := db.QueryRow("select gid from tPermissionGroups where memberPRI = ? and gid = ?", takerPRI, validGroupID)

		var w string
		err := row.Scan(&w)
		if err == sql.ErrNoRows {
			existed = false
		} else if err != nil {
			return err
		}

		if !existed {
			_, err := db.Exec("insert into tPermissionGroups(memberPRI, gid) values(?, ?)", takerPRI, validGroupID)
			if err != nil {
				return err
			}
		}

	}

	// clean sys group
	_, err = db.Exec("delete from tPermissionGroups where memberPRI = ? and gid != ?", takerPRI, validGroupID)
	if err != nil {
		return err
	}

	return nil
}
