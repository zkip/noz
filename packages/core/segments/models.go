package main

import (
	"fmt"

	"noz.zkip.cc/utils"
	"noz.zkip.cc/utils/model"
)

type R = model.Result

type setterImageAliasOption struct {
	ID   string
	Name string
}

type setterPaperOption struct {
	ID      string
	Name    string
	Content string
}

type PaperMetaData struct {
	Name string
	ID   string
}

type ImageMetaData struct {
	Name string
	ID   string
}

type PaperData struct {
	Name    string
	ID      string
	Content []byte
}

type PaperDataOption struct {
	ID      string
	Content string
}

type PaperMetaListResult struct {
	Data []*PaperMetaData
}

type ImageMetaListResult struct {
	Data []*ImageMetaData
}

type PaperListResult struct {
	Data []PaperData
}

type AccountResult struct {
	ID       string
	Nickname string
	Email    string
}

type AccountPatchOption struct {
	ID       string
	Nickname string
	Email    string
}

type Quota struct {
	Capcity uint64
	Used    uint64
}

func (q *Quota) doDelta(size int64) error {
	used := int64(q.Used) + size

	if used+size < 0 {
		used = 0
	}

	if used+size > int64(q.Capcity) {
		return &QuotaLackErr{q}
	}

	q.Used = uint64(used)

	return nil
}

type QuotaResult struct {
	Capcity uint64
	Used    uint64
}

type NoStoreResourceTypeErr struct {
	rType uint8
}

func (ne *NoStoreResourceTypeErr) Error() string {
	return fmt.Sprintf("%s is no store resource type.", utils.GetResoureceTypeIdent(ne.rType))
}

type QuotaLackErr struct {
	quota *Quota
}

func (qe *QuotaLackErr) Error() string {
	return fmt.Sprintf("Quota is lacked (%d/%d)).", qe.quota.Capcity, qe.quota.Used)
}

type UnsafeMoveErr struct{}

func (ue *UnsafeMoveErr) Error() string {
	return "Unsafe move."
}

type HierarchyRecord struct {
	Size  uint
	Order uint
	ID    string
	Name  string
	Path  []string
}

type HierarchyRecordListResult map[string]*HierarchyRecord

type HierarchyRecordOption struct {
	ID   string
	Name string
}
type HierarchyRecordMoverOption struct {
	ID       string
	ParentID string
	Order    uint
}
type HierarchyRecordDeleteBack struct {
	Name     string
	ParentID string
}
