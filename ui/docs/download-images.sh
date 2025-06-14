#!/bin/bash

# Create images directory
mkdir -p images

# Download images with descriptive names
curl -L "https://i.imgur.com/G5g2BwL.png" -o "images/elevation-scale.png"
curl -L "https://i.imgur.com/nJTRJ4h.png" -o "images/hexabase-logo-white.png"
curl -L "https://i.imgur.com/L4xK5JZ.png" -o "images/hexabase-logo-black.png"
curl -L "https://i.imgur.com/hOa48ay.png" -o "images/hexabase-logo-vertical-white.png"
curl -L "https://i.imgur.com/q2yX9qS.png" -o "images/hexabase-logo-vertical-black.png"
curl -L "https://i.imgur.com/2s0v5L9.png" -o "images/button-component-states.png"
curl -L "https://i.imgur.com/iC5Jv7q.png" -o "images/input-field-component-states.png"
curl -L "https://i.imgur.com/FjI9S6b.png" -o "images/tag-component-examples.png"
curl -L "https://i.imgur.com/F6yFz6E.png" -o "images/dialog-component-examples.png"
curl -L "https://i.imgur.com/2sPPCmU.png" -o "images/modal-component-examples.png"
curl -L "https://i.imgur.com/rW7V93q.png" -o "images/modal-padding-specification.png"
curl -L "https://i.imgur.com/N1sYpYy.png" -o "images/example-icons.png"

echo "All images downloaded successfully!" 