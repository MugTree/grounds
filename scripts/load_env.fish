# for line in (grep -v '^#' .env)
#     set parts (string split -m 1 '=' $line)
#     set -x $parts[1] $parts[2]
# end