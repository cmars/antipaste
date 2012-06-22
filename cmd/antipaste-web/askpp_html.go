package antipaste

import (
	"html/template"
)

var askppTemplate *template.Template

func init() {
	askppTemplate = template.Must(template.New("askpp").Parse(
`<HTML><HEAD>` + HEAD + `</HEAD><BODY>` + HEADER + `
<H2>Passphrase Required</H2>
<P>Your passphrase is needed in order to decrypt this paste.</P>
<P>If you are the only user of this computer and you're sure no one 
else is logged in, it should be safe for you to submit the passphrase here.</P>
<P>However, consider installing gpg-agent to manage your passphrase instead.</P>
<FORM NAME="askpp" METHOD="POST" ACTION="/pp">
<DIV id="askpp-section">
<INPUT type="password" name="pp"></INPUT>
<INPUT type="hidden" name="p" value="{{.Url}}></INPUT>
<INPUT type="submit" value="Submit"></INPUT>
</DIV>
</FORM>
` + FOOTER + `</BODY></HTML>`))
}

type askppArgs struct {
	PageName string
	Url string
}
