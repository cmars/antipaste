package antipaste

import (
	"html/template"
)

var indexTemplate *template.Template

func init() {
	indexTemplate = template.Must(template.New("new").Parse(
`<HTML><HEAD>` + HEAD +
`<script>
	$(function() {
		var recipients = [];
{{range .Identities}}
		recipients[recipients.length] = {label:"{{.Name}}", value:"{{.Fingerprint}}"};
{{end}}
		var acconfig = {
			source: recipients,
			select: function(event, ui) {
				updateInputs();
			}
		};
		var updateInputs = function(){
			$(".recipient").each(function(index, elt){
				if (elt.value == "") {
					elt.parentElement.removeChild(elt);
				}
			});
			var newRecipient = $('<INPUT type="text" name="recipient" class="recipient" size=40>');
			newRecipient.autocomplete(acconfig);
			newRecipient.change(updateInputs);
			newRecipient.focusout(updateInputs);
			$("#recipient-select").append(newRecipient);
		};
		updateInputs();
	});
</script>` +
`</HEAD><BODY>` + HEADER + `
<FORM id="open-form" NAME="open" METHOD="GET" ACTION="/">
<DIV class="container">
<DIV class="span-6">
<H2>Decrypt Paste</H2>
</DIV>
<DIV class="span-18 last">
<H3>Public URL</H3>
<INPUT type="text" name="p" class="index">
<INPUT type="submit" value="Decrypt" class="index">
</DIV>
</DIV>
</FORM>
<FORM NAME="create" METHOD="POST" ACTION="/paste">
<DIV class="container">
<DIV class="span-6">
<H2>Encrypt Paste</H2>
</DIV>
<DIV class="span-18 last">
<H3>To</H3>
<DIV id="recipient-select">
<INPUT type="text" id="recipient-first" name="recipient" class="recipient">
</DIV>
</DIV>
<DIV class="span-24 last">
<TEXTAREA id="contents" name="contents"></TEXTAREA>
</DIV>
<DIV class="span-24 last" id="submit-buttons">
<INPUT type="submit" value="Encrypt">
</DIV>
</DIV>
</FORM>
` + FOOTER + `</DIV></BODY></HTML>`))
}

type indexArgs struct {
	PageName string
	Identities []*Identity
}
