# yq-go eats up blank lines
# this function gets around that using diff with --ignore-blank-lines
yq_ibl()
{
  error=0
  diff_out=$(diff -B <(cat "$2") <(yq "$1" "$2")) || error=$?
  if [ "$error" != "0" ] && [ "$error" != "1" ]; then
    exit "$error"
  fi
  if [ -n "$diff_out" ]; then
    echo "$diff_out" | patch --quiet --no-backup-if-mismatch "$2" -
  fi
}
