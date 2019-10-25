#!/bin/sh -ex

DIR=`dirname "$0"`
RSC_FILE="../resources.go"
PKG="ui"
fyne=`go env GOPATH`/bin/fyne

cd $DIR
rm -f $RSC_FILE

append=""
for file in *.svg
do
  name=`basename $file .svg | perl -pe 's#(^|-|_)([a-z])#\u$2#g'`
  test -z "$append" && {
    $fyne bundle -package $PKG -name rsc$name $file > $RSC_FILE
  } || {
    $fyne bundle -package $PKG -name rsc$name -append $file >> $RSC_FILE
  }
  append="yes"
done
