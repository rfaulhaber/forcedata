package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rfaulhaber/forcedata/auth"
	"github.com/rfaulhaber/forcedata/job"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type CtxFlags struct {
	ObjFlag    string
	DelimFlag  string
	WatchFlag  time.Duration
	InsertFlag bool
	UpdateFlag bool
	UpsertFlag bool
	DeleteFlag bool
}

// context specifies the context in which the program runs
type Context struct {
	// "insert", "update", "upsert", or "delete"
	command  string
	flags    CtxFlags
	job      *job.Job
	setWatch bool
}

func NewRunContext(command string, flags CtxFlags, session auth.Session) (*Context, error) {
	delimName, ok := job.GetDelimName(flags.DelimFlag)

	if !ok {
		return nil, errors.Errorf("Invalid delimiter: %s", flags.DelimFlag)
	}

	config := job.JobConfig{
		Object:      flags.ObjFlag,
		Operation:   command,
		Delim:       delimName,
		ContentType: "CSV",
	}

	// TODO create job for each file

	return &Context{
		command,
		flags,
		job.NewJob(config, session),
		false,
	}, nil
}

func (ctx *Context) SetWatch() {
	ctx.setWatch = true
}

func (ctx *Context) Run(args []string) error {
	log.Println("running context")
	log.Println("creating job...")

	err := ctx.job.Create()

	if err != nil {
		return errors.Wrap(err, "ctx error creating job")
	}

	log.Println("uploading content...")
	var content []byte

	if isPipe() {
		content, err = readSource(os.Stdin)
	} else {
		// we only support uploading one file at a time at the moment!
		content, err = ioutil.ReadFile(args[0])
	}

	if err != nil {
		return errors.Wrap(err, "ctx could not read source")
	}

	err = ctx.job.Upload(content)

	if err != nil {
		return errors.Wrap(err, "ctx error uploading to job")
	}

	if ctx.setWatch {
		go ctx.job.Watch(ctx.flags.WatchFlag)

		for {
			select {
			case status, ok := <-ctx.job.Status:
				if ok {
					printStatus(status, os.Stdout)
				} else {
					return nil
				}
			case err := <-ctx.job.Error:
				return errors.Wrap(err, "watching job reported error")
			}
		}
	}

	return nil
}

func printStatus(status job.JobInfo, out io.Writer) {
	io.WriteString(out, "")
	fmt.Fprintf(out, "Records processed: %d\tRecords failed: %d", status.RecordsProcessed, status.RecordsFailed)
}

// reads content from source
func readSource(source io.ReadCloser) ([]byte, error) {
	content, err := ioutil.ReadAll(source)

	if err != nil {
		return nil, errors.Wrap(err, "attempting to read file source")
	}

	return content, nil
}

func isPipe() bool {
	stat, err := os.Stdin.Stat()
	return err == nil && stat.Size() > 0
}
