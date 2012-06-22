package antipaste

import (
	"html/template"
)

var showTemplate *template.Template

func init() {
	showTemplate = template.Must(template.New("show").Parse(
`<HTML><HEAD>` + HEAD + `</HEAD><BODY>` + HEADER + `
<DIV class="container">
<DIV id="show-title" class="span-24 last">
<H2>Decrypted Paste</H2>
<DIV id="show-desc">
<P>From public source <A target="_" href="{{.Url}}">{{.Url}}</A></P>
</DIV>
</DIV>
<DIV class="span-24 last">
<TEXTAREA id="contents" READONLY>{{.Paste}}</TEXTAREA>
</DIV>
</DIV>
` + FOOTER + `</BODY></HTML>`))
}

type showArgs struct {
	PageName string
	Url string
	Paste string
}
