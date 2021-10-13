#!/bin/bash

# workspace pwd
cd $(dirname $0)
cd "../"

if [ -z "$1" ]; then
	echo "Must specify a name"
	exit 1
fi

# starts with a capital letter
ComponentName=$(echo $1 | sed 's/\b[a-z]/\U&/g')

ComponentPath=src/components/$ComponentName

mkdir -p $ComponentPath

sed "s/\${NAME}/$ComponentName/g" scripts/template/entrypoint.ts.template > $ComponentPath/index.ts
sed "s/\${NAME}/$ComponentName/g" scripts/template/module.tsx.template > $ComponentPath/$ComponentName.tsx
cat scripts/template/module.scss.template > $ComponentPath/$ComponentName.module.scss
cat scripts/template/module.scss.d.ts.template > $ComponentPath/$ComponentName.module.scss.d.ts