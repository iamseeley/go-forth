---
title: 'Go-Forth 2.0'
description: 'Post 1 this is.'
---

**A static site generator written in Go.**

### Getting Started

- Clone the repository
- Make sure you have Go installed 
- When you run 'go build' dependencies will automatically be installed
- To build the static site run 'go build ./build/cmd/main.go' then './main build'
- To start the dev server run 'go build ./build/cmd/main.go' then './main dev'

### Add Content

Add Markdown files to the content directory. Folders within the content directory are collections of content. The page collection is special because markdown files placed here resolve to /$filename. You should place your index.md file here, which will resolve to the route of your site. 

### Create Collections

Want to create a new collection? Just add a folder to the content directory. If you are running the dev server, a corresponding template will be created for the collection in the template directory. 

### Display Data

If you want to display some data add a json file with some objects to the data directory. Then in the template you would like to display the data add a {{ range .Data.$filename }}
                {{ .$key }}
                {{ .$key }}
        {{ end }} to display a list of your data.

### Create Themes

Add css files(themes) and then point to the one you want to use in the config file. When you build the site the theme will get copied to the assets' css directory. You can add your other static assets like images and js files to the assets directory. Also, add your favicon to the assets directory!

### Edit & Create Templates

Add html templates to the templates directory to provide some structure to your content. This project uses Go's html/template package. I recommend reading the documentation to learn more about its templating features.

### Production

When you build the site for production all the necessary files will be placed in the src directory. This is the directory you will want to point to when hosting your site.

### Housekeeping

The routes will have the .html extension. If your hosting provider has a way to clean the url, I recommend doing that or you will need to add the .html extension to all the routes in your project. I'm working on a way to fix this. The dev server is setup so that routes resolve without the .html file extension.

### Thanks to

Russ Ross's [blackfriday](https://github.com/russross/blackfriday) and Go's standard library!