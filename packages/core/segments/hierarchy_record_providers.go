package main

import (
	"fmt"
	"net/http"

	"noz.zkip.cc/utils"
)

// hierarchy record
func hierarchyRecordListProvider(wr http.ResponseWriter, r *http.Request) {
	hu := utils.GenHandlerUtils(wr)
	db := utils.GetMySqlDB()

	userID := utils.ExtractUserIDFromRequst(r)
	userPRI := utils.GenPRI(userID, utils.Resource_type_user)

	var hrr = make(HierarchyRecordListResult)

	condQtr := "tHierarchy t on d.targetPRI = ? and t.descendant = d.hierarchyID order by t.descendant, t.distance desc"
	qtr := fmt.Sprintf("select d.hierarchyID, d.size, d.order, d.name, t.ancestor from tHierarchyData d inner join %s", condQtr)
	rows, err := db.Query(qtr, userPRI)
	utils.PanicIfErr(err)

	var hpath = []string{}
	for rows.Next() {
		var id, name, ancestor string
		var size, order uint
		err := rows.Scan(&id, &size, &order, &name, &ancestor)
		utils.PanicIfErr(err)

		if id == ancestor {
			hrr[id] = &HierarchyRecord{
				ID:    id,
				Name:  name,
				Size:  size,
				Order: order,
				Path:  hpath,
			}
			hpath = []string{}
		} else {
			hpath = append(hpath, ancestor)
		}
	}

	hu.SendJSON(0, "", hrr)
}

func hierarchyRecordRemoverProvider(wr http.ResponseWriter, r *http.Request) {
	hu := utils.GenHandlerUtils(wr)

	id := utils.ExtractDynamicRouteID(dynamic_route_pattern_hierarchy_record, r)

	// check args
	if utils.RunIfOK(id == "", hu.SendOptionErr, hu.ByErrJSON(37, "")) {
		return
	}

	_, err := deleteHierarchyRecord(id)
	utils.PanicIfErr(err)
}

func hierarchyRecordMoverProvider(wr http.ResponseWriter, r *http.Request) {
	hu := utils.GenHandlerUtils(wr)

	var option *HierarchyRecordMoverOption
	if utils.RunIfErr(utils.ExtractJsonBodyFromRequest(r, &option), hu.SendIllegalJsonOptionErr) {
		return
	}

	// check args
	if utils.RunIfOK(option.ID == "" || option.ParentID == "", hu.SendOptionErr) {
		return
	}

	err := moevHierarchyRecord(option.ID, option.ParentID, option.Order)
	if utils.RunIfPickErr(err, &UnsafeMoveErr{})(hu.ByErrJSONEM(72)) {
		return
	}
	utils.PanicIfErr(err)
}

func hierarchyRecordNameSetterProvider(wr http.ResponseWriter, r *http.Request) {

}

func hierarchyRecordFactoryProvider(wr http.ResponseWriter, r *http.Request) {

}
