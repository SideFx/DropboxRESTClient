// ---------------------------------------------------------------------------------------------------------------------
// (w) 2024 by Jan Buchholz
// Data model, using Unison library (c) Richard A. Wilkes
// https://github.com/richardwilkes/unison
// ---------------------------------------------------------------------------------------------------------------------

package models

import (
	"Dropbox_REST_Client/api"
	"Dropbox_REST_Client/assets"
	"Dropbox_REST_Client/dialogs"
	"fmt"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/fatal"
	"github.com/richardwilkes/toolbox/tid"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"math"
	"path"
	"slices"
	"strings"
)

// ---------------------------------------------------------------------------------------------------------------------
// FileSystem model
// ---------------------------------------------------------------------------------------------------------------------

const dragKey = "fileSystemRow"

var _ unison.TableRowData[*fileSystemRow] = &fileSystemRow{}
var fileSystemTable *unison.Table[*fileSystemRow]
var selectedRows []*fileSystemRow

type Caption struct {
	Title string
	Align align.Enum
}

type TableHeaderDescription struct {
	NoOfColumns int
	Captions    []Caption
}

type fileSystemItem struct {
	Name     string
	DbxId    string
	Modified string
	Size     string
	Hash     string
	Path     string
	IsFolder bool
}

type fileSystemRow struct {
	table        *unison.Table[*fileSystemRow]
	parent       *fileSystemRow
	children     []*fileSystemRow
	container    bool
	open         bool
	doubleHeight bool
	id           tid.TID
	_parent      *fileSystemRow
	M            fileSystemItem
}

var fileSystemTableDescription = TableHeaderDescription{
	NoOfColumns: 6,
	Captions: []Caption{
		{assets.CapName, align.Start},
		{assets.CapId, align.Start},
		{assets.CapModified, align.Start},
		{assets.CapSize, align.End},
		{assets.CapHash, align.Start},
		{assets.CapPath, align.Start},
	},
}

func NewFileSystemTable() (*unison.Table[*fileSystemRow], *unison.TableHeader[*fileSystemRow]) {
	unison.DefaultTableTheme.IndirectSelectionInk = unison.DefaultTableTheme.BackgroundInk
	unison.DefaultTableTheme.OnIndirectSelectionInk = unison.DefaultTableTheme.OnBackgroundInk
	unison.DefaultTableTheme.BandingInk = unison.DefaultTableTheme.BackgroundInk
	unison.DefaultTableTheme.OnBandingInk = unison.DefaultTableTheme.OnBackgroundInk
	fileSystemTable = unison.NewTable[*fileSystemRow](&unison.SimpleTableModel[*fileSystemRow]{})
	fileSystemTable.Columns = make([]unison.ColumnInfo, fileSystemTableDescription.NoOfColumns)
	fileSystemTable.HierarchyColumnID = 0
	for i := range fileSystemTable.Columns {
		fileSystemTable.Columns[i].ID = i
		fileSystemTable.Columns[i].Minimum = 70
		fileSystemTable.Columns[i].Maximum = 1000
	}
	fileSystemTable.SizeColumnsToFit(true)
	fileSystemTableHeader := unison.NewTableHeader[*fileSystemRow](fileSystemTable,
		unison.NewTableColumnHeader[*fileSystemRow](fileSystemTableDescription.Captions[0].Title, ""),
		unison.NewTableColumnHeader[*fileSystemRow](fileSystemTableDescription.Captions[1].Title, ""),
		unison.NewTableColumnHeader[*fileSystemRow](fileSystemTableDescription.Captions[2].Title, ""),
		unison.NewTableColumnHeader[*fileSystemRow](fileSystemTableDescription.Captions[3].Title, ""),
		unison.NewTableColumnHeader[*fileSystemRow](fileSystemTableDescription.Captions[4].Title, ""),
		unison.NewTableColumnHeader[*fileSystemRow](fileSystemTableDescription.Captions[5].Title, ""),
	)
	fileSystemTable.InstallDragSupport(nil, dragKey, "Row", "Rows")
	unison.InstallDropSupport[*fileSystemRow, any](fileSystemTable, dragKey,
		func(from, to *unison.Table[*fileSystemRow]) bool {
			return from == to
		},
		func(from, to *unison.Table[*fileSystemRow], move bool) *unison.UndoEdit[any] {
			selectedRows = nil // clear selection
			for _, row := range from.SelectedRows(true) {
				selectedRows = append(selectedRows, row) // copy selection to "var selectedRows []*fileSystemRow"
			}
			return nil
		},
		func(undo *unison.UndoEdit[any], from, to *unison.Table[*fileSystemRow], move bool) {
		},
	)

	fileSystemTable.DropOccurredCallback = func() {
		DropboxMoveFileItems() // perform move operation
	}
	return fileSystemTable, fileSystemTableHeader
}

func (d *fileSystemRow) CloneForTarget(target unison.Paneler, newParent *fileSystemRow) *fileSystemRow {
	table, ok := target.(*unison.Table[*fileSystemRow])
	if !ok {
		fatal.IfErr(errs.New("invalid target"))
	}
	clone := *d
	clone.table = table
	clone.parent = newParent
	clone._parent = newParent
	clone.id = tid.MustNewTID('a')
	return &clone
}

func (d *fileSystemRow) ID() tid.TID {
	return d.id
}

func (d *fileSystemRow) Parent() *fileSystemRow {
	return d.parent
}

func (d *fileSystemRow) SetParent(parent *fileSystemRow) {
	d.parent = parent
}

func (d *fileSystemRow) CanHaveChildren() bool {
	return d.container
}

func (d *fileSystemRow) Children() []*fileSystemRow {
	return d.children
}

func (d *fileSystemRow) SetChildren(children []*fileSystemRow) {
	d.children = children
}

func (d *fileSystemRow) CellDataForSort(col int) string {
	switch col {
	case 0:
		return d.M.Name
	case 1:
		return d.M.DbxId
	case 2:
		return d.M.Modified
	case 3:
		return d.M.Size
	case 4:
		return d.M.Hash
	case 5:
		return d.M.Path
	default:
		return ""
	}
}

func (d *fileSystemRow) ColumnCell(_, col int, foreground, _ unison.Ink, _, _, _ bool) unison.Paneler {
	var text string
	switch col {
	case 0:
		text = d.M.Name
	case 1:
		text = d.M.DbxId
	case 2:
		text = d.M.Modified
	case 3:
		text = d.M.Size
	case 4:
		text = d.M.Hash
	case 5:
		text = d.M.Path
	default:
		text = ""
	}
	wrapper := unison.NewPanel()
	wrapper.SetLayout(&unison.FlexLayout{Columns: 1, HAlign: fileSystemTableDescription.Captions[col].Align})
	addText(wrapper, text, foreground, unison.LabelFont)
	return wrapper
}

func (d *fileSystemRow) IsOpen() bool {
	return d.open
}

func (d *fileSystemRow) SetOpen(open bool) {
	var children []*fileSystemRow
	d.open = open
	// chevron open, no children loaded
	if open && len(d.children) == 0 {
		entries, err := api.ListFolders(d.M.DbxId, false, 2000)
		if err == nil {
			for _, entry := range entries {
				row := newFileSystemRow(tid.MustNewTID('a'), *entry, d)
				children = append(children, row)
			}
			if len(children) > 0 {
				d.SetChildren(children)
			}
		}
	}
	fileSystemTable.SyncToModel()
	for i := 0; i < fileSystemTableDescription.NoOfColumns; i++ {
		fileSystemTable.SizeColumnToFit(i, true)
	}
}

func (d *fileSystemRow) DeleteChild(child *fileSystemRow) {
	for i, c := range d.children {
		if c == child {
			d.children = slices.Delete(d.children, i, i+1)
			return
		}
		i++
	}
}

func (d *fileSystemRow) AddChild(child *fileSystemRow) {
	d.children = slices.Insert(d.children, 0, child)
}

func newFileSystemRow(id tid.TID, data api.FileItemType, parent *fileSystemRow) *fileSystemRow {
	isFolder := data.Tag == api.DbxFolder
	row := &fileSystemRow{
		table:     fileSystemTable,
		id:        id,
		container: isFolder,
		open:      false,
		parent:    parent,
		children:  nil,
		_parent:   parent,
		M: fileSystemItem{
			data.Name,
			data.Id,
			convertTimestamp(data.ServerModified),
			convertBytes(data.Size),
			data.ContentHash,
			data.PathDisplay,
			data.Tag == api.DbxFolder},
	}
	return row
}

func DropboxReadRootFolders() {
	var rootfolders []*fileSystemRow
	entries, err := api.ListFolders("", false, 2000)
	if err == nil {
		for _, entry := range entries {
			row := newFileSystemRow(tid.MustNewTID('a'), *entry, nil)
			rootfolders = append(rootfolders, row)
		}
		if len(rootfolders) > 0 {
			fileSystemTable.SetRootRows(rootfolders)
			fileSystemTable.SelectByIndex(0)
		}
		fileSystemTable.SyncToModel()
		for i := 0; i < fileSystemTableDescription.NoOfColumns; i++ {
			fileSystemTable.SizeColumnToFit(i, true)
		}
	}
}

func DropboxMoveFileItems() {
	var fromPath, toPath string
	var err error
	var m *api.FileItemMetadataType
	for _, row := range selectedRows {
		m = nil
		if row._parent != nil { // parent before dnd
			fromPath = row._parent.M.Path
		} else {
			fromPath = api.DbxPathSeparator
		}
		if row.parent != nil { // parent after dnd
			toPath = row.parent.M.Path
		} else {
			toPath = api.DbxPathSeparator
		}
		if fromPath == toPath {
			continue
		}
		fromPath = path.Join(fromPath, row.M.Name)
		toPath = path.Join(toPath, row.M.Name)
		m, err = api.MoveFiles(fromPath, toPath)
		if err == nil {
			row.M.Path = m.Metadata.PathDisplay
			row.M.Modified = convertTimestamp(m.Metadata.ServerModified)
			row.M.Name = m.Metadata.Name
			row._parent = row.Parent()
			if row.parent != nil {
				row.parent.SetOpen(true) // expand new parent
			}
		} else {
			if row._parent != nil {
				row._parent.AddChild(row) // old parent (drag source)
			}
			if row.parent != nil {
				row.parent.DeleteChild(row) // new parent
			}
			row.parent = row._parent // restore parent
			break
		}
	}
	fileSystemTable.SyncToModel()
	selectedRows = nil
	if err != nil {
		dialogs.DialogToDisplaySystemError(assets.TxtDropboxError, err)
	}
}

func addText(parent *unison.Panel, text string, ink unison.Ink, font unison.Font) {
	tx := unison.NewText(text, &unison.TextDecoration{Font: font})
	label := unison.NewLabel()
	label.Font = font
	label.LabelTheme.OnBackgroundInk = ink
	label.SetTitle(tx.String())
	parent.AddChild(label)
}

func convertBytes(b int64) string {
	if b == 0 {
		return ""
	}
	bf := float64(b)
	for _, unit := range []string{"", "k", "M", "G", "T"} {
		if math.Abs(bf) < 1024.0 {
			return fmt.Sprintf("%3.1f%sB", bf, unit)
		}
		bf /= 1024.0
	}
	return fmt.Sprintf("%.1fYiB", bf)
}

func convertTimestamp(timestamp string) string {
	result := strings.Replace(timestamp, "T", " ", 1)
	return strings.Replace(result, "Z", " ", 1)
}
