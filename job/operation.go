package job

type Operation int

const (
	Insert = iota
	Update
	Delete
	Upsert
)

var operations = []string{
	"insert",
	"update",
	"delete",
	"upsert",
}

var operationMap = map[string]Operation{
	"insert": Insert,
	"update": Update,
	"delete": Delete,
	"upsert": Upsert,
}

func (o Operation) String() string {
	return operations[o]
}

func isOperation(str string) bool {
	_, ok := operationMap[str]
	return ok
}
