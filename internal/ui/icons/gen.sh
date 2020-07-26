#!/bin/sh -ex

DIR=$(dirname "$0")
STATIC_FILE="../generated_icons_static.go"
THEMED_FILE="../generated_icons_themed.go"
PKG="ui"
fyne=$(go env GOPATH)/bin/fyne

cd "$DIR"
rm -f $STATIC_FILE
rm -f $THEMED_FILE

append=""
for file in *.svg
do
  name=$(basename "$file" .svg | perl -pe 's#(^|-|_)([a-z])#\u$2#g')
  if test -z "$append"; then
    $fyne bundle -package $PKG -name "rsc$name" "$file" > $STATIC_FILE
    {
      echo "// auto-generated"
      echo
      echo "package $PKG"
      echo
      echo 'import "fyne.io/fyne/theme"'
    } > $THEMED_FILE
  else
    $fyne bundle -package $PKG -name "rsc$name" -append "$file" >> $STATIC_FILE
  fi
  {
    echo
    echo "// ${name}Icon is a themed version of the ${name} resource."
    echo "var ${name}Icon *theme.ThemedResource"
    echo
    echo "func init() {"
    printf "\t%sIcon = theme.NewThemedResource(rsc%s, nil)\n" "$name" "$name"
    echo "}"
  } >> $THEMED_FILE
  append="yes"
done
