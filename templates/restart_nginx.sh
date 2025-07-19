#!/bin/bash
echo "Restarting nginx..."
systemctl restart nginx
systemctl status nginx | head -n 5
