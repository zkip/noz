package main

import (
	"net/http"

	"noz.zkip.cc/utils"
)

var maxUploadSize int64 = 1024 * 1024 * 1024 * 1024

const (
	dynamic_route_pattern_image            = `^/image/([^\/]+)$`
	dynamic_route_pattern_paper            = `^/paper/([^\/]+)$`
	dynamic_route_pattern_hierarchy_record = `^/hierarchy_record/([^\/]+)$`
)

var resource_store_path = map[uint8]string{
	utils.Resource_type_image: "data/images",
	utils.Resource_type_paper: "data/papers",
}

var (
	ResourceNonExistErr_E = &ResourceNonExistErr{}
)

func setupServer() {
	httpServer := utils.NewNOZHTTPServer()

	httpServer.HandleFunc(imageFactoryProvider, "^/image", utils.Http_method_new)
	httpServer.HandleFunc(imageProvider, "^/image", utils.Http_method_get)
	httpServer.HandleFunc(imageRemoverProvider, "^/image", utils.Http_method_delete)
	httpServer.HandleFunc(imageAliasSetterProvider, "^/image/alias", utils.Http_method_set)
	httpServer.HandleFunc(imageListProvider, "^/image/list", utils.Http_method_action)

	httpServer.HandleFunc(paperFactoryProvider, "^/paper", utils.Http_method_new)
	httpServer.HandleFunc(paperProvider, "^/paper", utils.Http_method_get)
	httpServer.HandleFunc(paperRemoverProvider, "^/paper", utils.Http_method_delete)
	httpServer.HandleFunc(paperListProvider, "^/paper/list", utils.Http_method_action)

	httpServer.HandleFunc(paperAliasSetterProvider, "^/paper/name", utils.Http_method_set)
	httpServer.HandleFunc(paperContentSetterProvider, "^/paper/content", utils.Http_method_set)

	httpServer.HandleFunc(hierarchyRecordRemoverProvider, "^/hierarchy_record", utils.Http_method_delete)
	httpServer.HandleFunc(hierarchyRecordNameSetterProvider, "^/hierarchy_record/name", utils.Http_method_set)
	httpServer.HandleFunc(hierarchyRecordFactoryProvider, "^/hierarchy_record", utils.Http_method_new)
	httpServer.HandleFunc(hierarchyRecordMoverProvider, "^/hierarchy_record/move", utils.Http_method_action)
	httpServer.HandleFunc(hierarchyRecordListProvider, "^/hierarchy_record/list", utils.Http_method_action)

	httpServer.HandleFunc(quotaProvider, "^/quota", utils.Http_method_get)
	httpServer.HandleFunc(accountProvider, "^/account", utils.Http_method_get)
	httpServer.HandleFunc(accountPatcherProvider, "^/account", utils.Http_method_patch)

	utils.PanicIfErr(http.ListenAndServe(utils.GetServeHost(7703), httpServer))
}

func main() {
	defer utils.Setup()()

	test()

	setupStoreEnv()
	setupServer()
}

func test() {

	// var err error
	// ownerPRI := "us/1"
	// _, err = newHierarchyRecord("", ownerPRI, "crack")
	// utils.PanicIfErr(err)

	/*
		crack
			plan A
				crash stock
					xxx
				sky mesh
				field jump
			plan B
				run
	*/

	db := utils.GetMySqlDB()

	db.Exec("truncate tHierarchy")
	db.Exec("truncate tHierarchyData")

	ownerPRI := "us/1"
	idRoot, _ := insertHierarchyRecord("", ownerPRI, "crack")
	idPlanA, _ := insertHierarchyRecord(idRoot, ownerPRI, "plan A")
	idStock, _ := insertHierarchyRecord(idPlanA, ownerPRI, "crash stock")
	insertHierarchyRecord(idStock, ownerPRI, "xxx")
	insertHierarchyRecord(idPlanA, ownerPRI, "field jump")
	insertHierarchyRecord(idPlanA, ownerPRI, "sky mesh", 1)
	idPlanB, _ := insertHierarchyRecord(idRoot, ownerPRI, "plan B")
	insertHierarchyRecord(idPlanB, ownerPRI, "run")

	// fmt.Println(findPath("20046eb7d5ca154469c4d07cd0de5b61"))

	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(id, "@@@@")

	// err := delete("955e46d9a93517908728b5d4046cfd27")
	// utils.PanicIfErr(err)

	// findDepth(2)
	// fmt.Println(findChildren("bda78e4a601297d7eb9aa6d608e801c2"))
	// findPath(3)

	// renameHierarchyRecord("bda78e4a601297d7eb9aa6d608e801c2", "Plan X")
}
