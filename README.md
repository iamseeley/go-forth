# Go-Forth 2.0

Go-Forth is a static site generator written in Go. 

![image](https://github.com/iamseeley/go-forth2.0/assets/104278845/82070d1c-052c-4470-b595-92717d4eb8fd)

## Getting Started

1. Clone the [repository](https://github.com/iamseeley/go-forth2.0)
2. Make sure you have Go installed
3. Run 'npm install' to install puppeteer (for og-image generation)
4. To start the dev server run 'go build ./build/cmd/main.go' then './main dev'
5. To build the static site run 'go build ./build/cmd/main.go' then './main build'

## Structure

<div>
<ul class="structure">
  <li>
    <details>
      <summary>go-forth2.0/</summary>
      <ul>
        <li>
          <details>
            <summary>assets/</summary>
            <ul>
              <li>css/</li>
              <li>images/</li>
              <li>js/</li>
              <li>og-image/</li>
              <li>favicon.ico</li>
            </ul>
          </details>
        </li>
        <li>build stuffs/</li>
        <li>
          <details>
            <summary>content/</summary>
            <ul>
              <li><a href="/">index.md</a></li>
              <li>
                <details>
                  <summary>post/</summary>
                  <ul>
                    <li><a href="/post/post1">post1.md</a></li>
                  </ul>
                </details>
              </li>
            </ul>
          </details>
        </li>
        <li>
          <details>
            <summary>data/</summary>
            <ul>
              <li>data.json</li>
            </ul>
          </details>
        </li>
        <li>src/ (output)</li>
        <li>
          <details>
            <summary>templates/</summary>
            <ul>
              <li>page.html</li>
              <li>og-image.html</li>
              <li>site.html</li>
            </ul>
          </details>
        </li>
        <li>
          <details>
            <summary>themes/</summary>
            <ul>
              <li>default.css</li>
            </ul>
          </details>
        </li>
        <li>config.json</li>
      </ul>
    </details>
  </li>
</ul>
</div>

## Add Content

Add Markdown files to the content directory. Folders within the content directory are collections of content. The page collection is special because markdown files placed here resolve to /$filename. You should place your index.md file here, which will resolve to the route of your site. 

### Create Collections

Want to create a new collection? Just add a folder to the content directory. Make sure to add a template with the same filename as the collection. If you are running the dev server and you add a new collection, a corresponding template will be created in the template directory. 

### Display Data

If you want to display some data add a json file with some objects to the data directory. Then in the template you would like to display the data add a {{ range .Data.$filename }}
                {{ .$key }}
                {{ .$key }}
        {{ end }} to display a list of your data.

## Create Themes

Add css files(themes) and then point to the one you want to use in the config file. When you build the site the theme will get copied to the assets' css directory. You can add your other static assets like images and js files to the assets directory. Also, add your favicon to the assets directory!

### Edit & Create Templates

Add html templates to the templates directory to provide some structure to your content. This project uses Go's html/template package. I recommend reading the documentation to learn more about its templating features.

## Production

When you build the site for production all the necessary files will be placed in the src directory. This is the directory you will want to point to when hosting your site.

## Housekeeping

The routes will have the .html extension. If your hosting provider has a way to clean the url, I recommend doing that or you will need to add the .html extension to all the routes in your project. I'm working on a way to fix this. The dev server is setup so that routes resolve without the .html file extension.

## Thanks to

Russ Ross's [blackfriday](https://github.com/russross/blackfriday) and Go's standard library!
