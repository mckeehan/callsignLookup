# callsignLookup

This little program takes the FCC callsign data fetched by Xastir (see https://github.com/Xastir/Xastir/blob/master/scripts/get-fcc-rac.pl.in) and insert it into a database where it can be quickly queried.

Its default invocation will query the database for each argument as a callsign.

The original implementation of this was a simple bash script that used grep and awk to lookup the callsign (from the same Xastir data file). This approach was very slow and lead me to a different approach.
```bash
#!/bin/bash
CALL_SIGN=$(echo "$1" | tr '[a-z]' '[A-Z]' )
# LC_ALL=C cat /usr/local/share/xastir/fcc/EN.dat | awk -F '|' '{ printf "%s|%s %s|%s|%s, %s\n", $5, $9, $11, $16, $17, $18 }' | fgrep -F "${CALL_SIGN}|" | tr '|' '\n'
RESULT=$(LC_ALL=C fgrep -F "${CALL_SIGN}|" /usr/local/share/xastir/fcc/EN.dat | awk -F '|' '{ printf "%s|%s %s|%s|%s, %s\n", $5, $9, $11, $16, $17, $18 }')
echo "$RESULT" | tr '|' '\n'
echo -n "http://maps.apple.com/?address="
echo "$RESULT" | awk -F'|' '{ print $3 "," $4 }' | sed 's/ /+/g'
```

The `-alfred` flag was added to support outputting queery results in a JSON format that is used by the [Alfred Script Filter](https://www.alfredapp.com/help/workflows/inputs/script-filter/json/).
