package meta_test

import (
	"testing"

	"github.com/asteris-llc/converge/resource/meta"
	"github.com/stretchr/testify/assert"
)

func TestMeta(t *testing.T) {
	m := &meta.Meta{Author: "sehqlr"}
	expected := `meta:
	Status:	{map[] [] no change []}
	Author:	sehqlr
	Organization:	
	PgpKeyId:	
	OrgUrl:	
	Version:	
	VcsUrl:	
	License:	
	VcsCommit:	
	Description:	`
	assert.Equal(t, expected, m.String())
}

/*
	metaValue := reflect.ValueOf(m).Elem()
	stringSlice := []string{"meta:"}

	for i := 0; i < metaValue.NumField(); i++ {
		key := metaValue.Type().Field(i).Name
		value := metaValue.Field(i)

		stringSlice = append(stringSlice, fmt.Sprintf("%v:\t%v", key, value))

	}
	metaString := strings.Join(stringSlice, "\n\t")

	fmt.Print(metaString)
*/
