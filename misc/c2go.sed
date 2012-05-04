# this sed script replaces some bits of the original leg file
# to make it more similar to the Go version, thus avoiding
# to many differences

/\$\$/ {
	s,\$\$->,$$.,g
	/\$\$[^}]*$/s,\; *$,,g
}

s,parse_result,p.tree,
s,references,p.references,
s,notes,p.notes,
s,find_reference,p.findReference,g

s,->key,.key,g
s,->children,.children,g
s,->contents.str,.contents.str,g

/{ *if (extens/ {
	s,if (,if ,
	s,)),),
}
/EXT/ s,if extension,if p.extension,
/EXT/ s,{ *extension,{ p.extension,g
/EXT/ s,{ *!extension,{ !p.extension,g
/EXT/ {
	s,extension.EXT_FILTER_HTML.,extension.FilterHTML,g
	s,extension.EXT_FILTER_STYLES.,extension.FilterStyles,g
	s,extension.EXT_SMART.,extension.Smart,g
	s,extension.EXT_NOTES.,extension.Notes,g
}

s,{ *element \*[a-z]*\; *$,{,

/raw\.key =/ s,;$,,
/result =/ s,;$,,
s,result = mk_el,result := mk_el,

s,NULL,nil,g

s, *\; *}, },g

s,strlen(,len(,g

s/mk_element/p.mkElem/
s/mk_str_from_list/p.mkStringFromList/
s/mk_str/p.mkString/g
s/mk_list/p.mkList/
s/mk_link/p.mkLink/
