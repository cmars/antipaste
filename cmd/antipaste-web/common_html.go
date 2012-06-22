package antipaste

const HEAD = `
<META http-equiv="Content-type" content="text/html; charset=utf-8" />
<TITLE>ANTI-PASTE | {{.PageName}}</TITLE>
<link type="text/css" href="/static/css/smoothness/jquery-ui-1.8.19.custom.css" rel="Stylesheet"></link>
<script type="text/javascript" src="/static/js/jquery-1.7.2.min.js"></script>
<script type="text/javascript" src="/static/js/jquery-ui-1.8.19.custom.min.js"></script>
<link href='http://fonts.googleapis.com/css?family=Francois+One|Ubuntu+Mono' rel='stylesheet' type='text/css'></link>
<link rel="stylesheet" href="/static/blueprint/screen.css" type="text/css" media="screen, projection" />
<link rel="stylesheet" href="/static/blueprint/print.css" type="text/css" media="print" />
<!--[if IE]><link rel="stylesheet" href="/static/blueprint/ie.css" type="text/css" media="screen, projection" /><![endif]-->
<!-- Import fancy-type plugin. -->
<link rel="stylesheet" href="/static/blueprint/plugins/fancy-type/screen.css" type="text/css" media="screen, projection" />
<STYLE>
H1, H2, H3 {
	font-family: 'Francois One', sans-serif;
	font-weight: 400;
	margin: 0px 0px 0px 0px;
	padding: 2px 2px 2px 2px;
}
H1 {
	font-size: 48pt;
	border-bottom: dotted 1px #ccc;
}
H2 {
	font-size: 28pt;
}
H3 {
	font-size: 18pt;
	vertical-align: bottom;
	padding: 0px 0px 0px 0px;
}
H1.logo a, H1.logo a:hover, H1.logo a:visited, H1.logo a:active {
	text-decoration: none;
	color: #000;
}
INPUT[type=text].index, INPUT[type=text].recipient, INPUT[type=submit].index {
	margin: 0;
}
TEXTAREA#contents {
	font-family: 'Ubuntu Mono', sans-serif;
	font-size: 14pt;
	font-weight: 400;
	width: 100%;
}
TEXTAREA#contents {
	height: 14em;
}
DIV#recipient-select INPUT[type=text] {
	display: block;
	font-size: 12pt;
	width: 28em;
}
FORM#open-form * {
	vertical-align: middle;
}
FORM#open-form INPUT {
	font-size: 12pt;
}
FORM#open-form INPUT[type=text] {
	width: 28em;
}
FORM#open-form H2 {
	vertical-align: top;
}
DIV#show-title * {
	display: table-cell;
	vertical-align: middle;
} 
DIV#show-desc {
	width: 28em;
}
DIV#show-desc P {
	padding-left: 0.5em;
	line-height: 95%;
	font-size: 14pt;
}
</STYLE>
`

const HEADER = `
<DIV class="container">
<DIV class="span-24 last">
<H1 class="logo"><a href="/">ANTI-PASTE</a></H1>
</DIV>
</DIV>
`

const FOOTER = `
<DIV class="container">
<DIV class="span-24 last">
<P class="footer">Visit the <A HREF="http://antipaste.com/">ANTI-PASTE</a> project website. Licensed for all under the Affero General Public License.</P>
</DIV>
</DIV>
`
