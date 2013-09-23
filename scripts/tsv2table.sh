#! /bin/bash

(cat `ls $1/*.tsv | head -1` | tail -n +17 | head -13 | grep -v "Response Time Details:" | ruby tsvhelper.rb | head -1 | nl -s File, | cut -c7-;

for n in `ls $1/*.tsv` ; do
  cat $n | tail -n +17 | head -13 | grep -v "Response Time Details:" | ruby tsvhelper.rb | tail -n +2 | nl -s $n, | cut -c7- ;
done) | column -t -s ','
