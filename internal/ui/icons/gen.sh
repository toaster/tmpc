#!/bin/sh -ex

DIR=`dirname "$0"`
STATIC_FILE="../icons_static.go"
THEMED_FILE="../icons_themed.go"
PKG="ui"
fyne=`go env GOPATH`/bin/fyne

cd $DIR
rm -f $STATIC_FILE
rm -f $THEMED_FILE

append=""
for file in *.svg
do
  name=`basename $file .svg | perl -pe 's#(^|-|_)([a-z])#\u$2#g'`
  test -z "$append" && {
    $fyne bundle -package $PKG -name rsc$name $file > $STATIC_FILE
    echo "package $PKG" > $THEMED_FILE
    echo >> $THEMED_FILE
    echo 'import "fyne.io/fyne/theme"' >> $THEMED_FILE
  } || {
    $fyne bundle -package $PKG -name rsc$name -append $file >> $STATIC_FILE
  }
  echo >> $THEMED_FILE
  echo "var ${name}Icon *theme.ThemedResource" >> $THEMED_FILE
  echo >> $THEMED_FILE
  echo "func init() {" >> $THEMED_FILE
  echo "\t${name}Icon = theme.NewThemedResource(rsc${name}, nil)" >> $THEMED_FILE
  echo "}" >> $THEMED_FILE
  append="yes"
done
