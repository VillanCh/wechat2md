package test

import (
	"fmt"
	"io/ioutil"

	"github.com/VillanCh/wechat2md/format"
	"github.com/VillanCh/wechat2md/parse"
)

func Test1() {
	var articleStruct parse.Article = parse.ParseFromHTMLFile("./test/test1.html", parse.IMAGE_POLICY_BASE64)
	fmt.Println("-------------------test1.html parse-------------------")
	fmt.Printf("%+v\n", articleStruct)

	fmt.Println("-------------------test1.html format-------------------")
	mdString, _ := format.Format(articleStruct)
	fmt.Print(mdString)
	ioutil.WriteFile("./test/test1_target.md", []byte(mdString), 0644)
}
