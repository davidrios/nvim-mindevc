#!/bin/sh
set -e

USERADD=$(ls /usr/sbin/useradd 2>/dev/null || echo "adduser -D")
test "$USERADD" != "/usr/sbin/useradd" || USERADD="$USERADD -m"
getent passwd user || $USERADD user

cat <<'EOF' > /usr/bin/ushell
#!/bin/sh
gosu="/opt/nvim-mindevc/bin/gosu"
grep -qs nvim-mindevc /home/user/.profile || $gosu user sh -c 'echo "export PATH=$PATH:/opt/nvim-mindevc/bin" >> ~/.profile'
exec $gosu user $(ls /bin/bash 2>/dev/null || echo sh) -il
EOF

chmod +x /usr/bin/ushell

sleep 365d
