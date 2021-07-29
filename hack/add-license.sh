#!/bin/sh
# script to copy the headers to all the source files and header files
for f in maasclient/*.go; do
  if (grep Copyright $f);then
    tail -n +17 $f > $f.new
    mv $f.new $f
    echo "Removed header"
  fi
    cat hack/apache.placeholder $f > $f.new
    mv $f.new $f
    echo "License Header copied to $f"
done