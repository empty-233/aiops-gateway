package logs

type SourceType string

const (
	SourceFile    SourceType = "file"
	SourceJournal SourceType = "journal"
)

type Source struct {
	Type SourceType
	Path string
}
