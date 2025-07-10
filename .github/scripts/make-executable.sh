#!/bin/bash
# Run this script to make all workflow scripts executable

chmod +x .github/scripts/*.sh

echo "All scripts are now executable!"
echo "Don't forget to commit the changes:"
echo "  git add .github/scripts/"
echo "  git commit -m 'Add executable workflow scripts'"