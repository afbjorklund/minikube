#!/bin/sh

# Copyright 2019 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

cd "$( dirname $0 )"

name=${1:-minikube}

# Create icons, from the original high-resolution logo image.
#
# * convert: sudo apt install imagemagick
# * svgo:    sudo npm install -g svgo

test -d logo || exit 1
mkdir -p icon || exit 1

logo=logo/logo.png
for size in 16 32 48 64 128 256; do
  icon=icon/${name}-${size}x${size}.png
  convert ${logo} -background none -resize ${size}x${size} \
                  -gravity center -extent ${size}x${size} \
                  -alpha on ${icon}
  file $icon
done

logo=logo/logo.svg
icon=icon/${name}-scalable.svg
svgo --quiet $logo -o $icon
sed -i '1i<?xml version="1.0" encoding="UTF-8" standalone="no"?>' $icon
file $icon
