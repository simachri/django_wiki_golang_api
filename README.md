# Django-Wiki API in Golang

  The API uses a Golang API server [Gin](https://github.com/gin-gonic/gin) and directly 
  updates the Postgres database through the Golang driver 
  [pgx](https://github.com/jackc/pgx). There is an alternative way to add API functions 
  to Django itself and call them from Golang API server instead of directly updating the 
  database.  The decision to directly interact with the database is on purpose to be 
  independent from the Django implementation and learn something about working 
  programmatically with a Postgres database.


## Open questions

 - [ ] Local files may be moved or renamed: How to remember which local file maps to 
   which article in the Wiki?

   - Always delete entire online Wiki before uploading?
     - Drawback: History is lost in online article.
     - Changes to the online article that are missing in the local file are missing.

   - Maintain a local database to store the mapping?

  - [ ] For each directory on the local machine, a page in the wiki needs to be created 
    to have the  URL paths correctly, see [here](#wiki_slug_dir). What is the content of 
    this page?


## Testing

### Spin up Docker containers for Django Wiki
<a id="launch_app"></a>

  ```sh
  $ cd ~/Development/Linode-coco-life/django-on-docker
  $ docker-compose down
  $ docker-compose up -d
  $ docker-compose logs
  ```

### Databes with test data

#### First time setup of test database

  1. [Spin up the Docker containers](#launch_app).
  1. Create a superuse account.
  1. Create a root article.
  1. Clone the database:
     1. Login to database `pk_db_dev` using [psql](#conn_to_pgdb).
     1. List existing databases: `\l`
     1. Run:
        - `create database go_api_tests with template pk_db_dev;`
        - `\q`
     1. Login to database `go_api_tests` using [psql](#conn_to_pgdb).
     1. Run `truncate wiki_article cascade;`
  
#### Create single articles

  Use `psql`:
  - `\set title 'New article'`
  - `\set content '# First level header'`
  - `\set slug 'foo'`
  - `\i test/sql/new_article.sql`

#### Clear all test data

  Use `psql`: `\i test/sql/delete_all_keep_root.sql`

  __Note__: `go test ...` - see below - programmatically clears the test database.


### Go tests

  From the project root run `go test -v ./...`.


### Interactively

  1. [Spin up the Docker containers](#launch_app).

  1. To connect to the PostgreSQL database from outside the Docker container use:
     1. Take the password from 
        `~/Development/Linode-coco-life/django-on-docker/.env.dev.db`
     1. `psql -U xi3k -h 0.0.0.0 -d pk_db_dev`
         <a id="conn_to_pgdb"></a>

  1. Maintain a file `.env`:
     ```
      PGHOST=0.0.0.0
      PGPORT=5432
      PGDATABASE=pk_db_dev
      PGUSER=xi3k
      PGPASSWORD=<DB password>
     ```

  1. Start the server:
     `go run ./...` 

  1. Test the following endpoints:
     - `localhost:8080/ping`
     - `localhost:8080/db/health`


## Data model of articles

  One article [wiki_article](#db_wiki_art) has _n_ (n >= 1) revisions 
  [wiki_articlerevision](#db_wiki_artrev).

  The actual content of an article is stored in [wiki_articlerevision](#db_wiki_artrev).

  The _URL assignment_ and the _hierarchy_ between articles is modeled in 
  [wiki_urlpath](#db_wiki_urlpath): To make an article accessible on the UI, it is 
  mandatory that is has a record in this table. Otherwise, the article can simply not be 
  shown as there is no dedicated URL.


## Database tables

### `wiki_article` - article header data
<a id="db_wiki_art"></a>

  ```
  -[ RECORD 1 ]-------+------------------------------
  id                  | 1
  created             | 2021-02-25 09:46:41.709591+00
  modified            | 2021-02-25 09:46:41.761749+00
  group_read          | t
  group_write         | t
  other_read          | t
  other_write         | t
  current_revision_id | 1
  group_id            |
  owner_id            |
  ```

  _Fields_:
  - `id`: Is automatically incremented by the database schema.
          <a id="wiki_art_id"></a>

  - `current_revision_id`: _Foreign key_ to [wiki_articlerevision-id](#wiki_artrev_id)

    __Notes__:
    - The `current_revision_id` is not available before the `wiki_articlerevision` record 
      has been created. Yet, the `wiki_articlerevision` record requires `wiki_article-id` 
      to be created. That is, `current_revision_id` needs to be filled _after_ 
      `wiki_articlerevision` has been created:
      1. Create new record in `wiki_article`.
      1. Create new record in `wiki_articlerevision` with foreign key to `wiki_article-id`.
      1. Update the record in `wiki_article` and set `current_revision_id`.

    - This __cannot__ be achieved using a trigger function on the database as this will 
      interfere with Django Wiki's database saving logic.

  - `created` and `modified`: For a reference, see where the timestamps are create in 
    class `Article`:
      <a id="wiki_art_crea"></a>

    ```python
      created = models.DateTimeField(
          auto_now_add=True,
          verbose_name=_("created"),
      )
      modified = models.DateTimeField(
          auto_now=True,
          verbose_name=_("modified"),
          help_text=_("Article properties last modified"),
      )
    ```

    [models/article.py](/home/xi3k/.local/share/virtualenvs/app-IosILp-G/lib/python3.9/site-packages/wiki/models/article.py)


### `wiki_articlerevision` - article content
<a id="db_wiki_artrev"></a>

  _Fields_:
  - `id`: Is automatically incremented by the database schema.
          <a id="wiki_artrev_id"></a>

  - `article_id`: _Foreign key_ to [wiki_article-id](#wiki_art_id)
  - `revision_number`: Starts at `1`.
  - `previous_revision_id`: Mandatory if `revision_number` is greather than `1`.
  - `content`: Contains the actual article markdown source
  - `title`: Contains the actual article markdown source
  - `created` and `modified`: See [wiki_article](#wiki_art_crea).
  - `deleted`: Has _notnull_ constraint.
  - `locked`: Has _notnull_ constraint.
  - `user_message`, `automatic_log`, `ip_address`: Have _notnull_ constraints but are 
    filled with empty string.


### `wiki_urlpath` - model hierarchy of articles in the wiki
<a id="db_wiki_urlpath"></a>

  _Fields_:
  - `id`: Is automatically incremented by the database schema.
  - `slug`:
    <a id="db_wiki_fld_slug"></a>
    - URL part identifying the current hierarchy level in the wiki without leading and 
      trailing `/` 
    - Is `null` for the __root node__.
    - The resulting URL is `https://<domain>/<parent>/../<parent>/<slug>`.
    - It __cannot__ contain a hierarchy, that is, a `/`.

  - `lft`:
  - `rght`:
      <a id="db_wiki_lftright_algo"></a>

    - Django Wiki uses a Django module
      [MPTT](https://django-mptt.readthedocs.io/en/latest/tutorial.html) which facilitates 
      storing hierarchical/tree data (in this case its the relation of the Wiki pages) by 
      implementing the _Modified Preorder Tree Traversal algorithm_. 

      - Guide inlcuding SQL queries on how to work with the data 
        [here](https://www.sitepoint.com/hierarchical-data-database-2).
      - Another good guide 
        [here](https://www.ibase.ru/files/articles/programming/dbmstrees/sqltrees.html)

    - __Algorithms__:
      - _Insert_ a new article `n` as _single child_ (new leaf) or _right sibling_ to the 
        rightmose direct child of _parent_ `p`:

        1. Set `lft` and `rght` of `n` based on `l`:
           - `n.lft = l.rght`
           - `n.rght = l.rght + 1`

        1. Adjust `lft` and `rght` of all nodes `r` that are
           - __either__ _right siblings_ to `n` (including their children)
           - __or__ _direct children_ of `n`
           - __or__ _direct parent_
           - __or__ _ancestors_ (parent and grandparent of _parent_)
           - __or__ _right_ to _direct parent_ or _ancestors_.

           All their `lft` and `rght` values need to be incremented by `2`:
           - `lft`: All nodes `r` with `r.lft >= n.lft`:
             `r.lft = r.lft + 2`

           - `rght`: All nodes `r` with `r.rght >= n.lft`:
             `r.rght = r.rght + 2`

              __Note__: The condition `r.rght >= r.rght` (mind the `rght` instead of the 
              `lft`) does not cover for parent nodes as their `r.rght` is not matched by 
              this condition. Example: Parent node has `r.lft = 1 and r.rght = 2`. New 
              node is inserted with `n.lft = 2 and n.rght = 3`. `r.rght` has to be set to 
              `4`.

      - _Delete_ an article `d`:
        - [ ] Review!
        All nodes `r` in the tree right to the deleted node (independent of the hierarchy 
        as it is modeled through field `parent_id`), need to have their `lft` and 
        `rght` properties decremented by `2` as there is one node missing now.
        1. _For all nodes with_ `r.rght > d.rght`:  `r.lft = r.lft - 2`.
        1. _For all nodes with_ `r.lft > d.rght`:  `r.rght = r.rght - 2`.

  - `level`:
    <a id="db_wiki_fld_level"></a>
    - Identifies the level in the tree hierarchy.
    - `0` is the level of the __root node__

  - `tree_id`:
    - Is defined by the root node and inherited by all its children.
    - Is `1` in our case.
    - [django-mptt](https://django-mptt.readthedocs.io/en/latest/technical_details.html#tree-id)

  - `article_id`:
    - _Foreign key_ to [wiki_article-id](#wiki_art_id)
    - The article that is displayed when opening the URL path.

  - `site_id`: Always `1`
  - `parent_id`: _Foreign key_ to the parent's `wiki_urlpath-id`


## Create new article

  Assumption(s):
  - Article does not yet exist.

  1. Fill [wiki_article](#db_wiki_art):

     ```sql
     insert into wiki_article (current_revision_id)
      values (1)
      returning id;
     ```

  1. Use `wiki_article-id` as `artid`.
  1. Fill [wiki_articlerevision](#db_wiki_art):

     ```sql
     insert into wiki_articlerevision (article_id, revision_number, title, content)
      values (<artid>, 1, <title>, <markdown content>)
      returning id;
     ```

  1. Fill [wiki_urlpath](#db_wiki_urlpath):

     ```sql
     insert into wiki_urlpath (slug, lft, rght, level, tree_id, article_id, site_id, parent_id)
      values (<slug>, <lft>, <rght>, <level>, 1, <artid>, 1, <parentid>)
      returning id;
     ```
     - [slug](#db_wiki_fld_slug)
       - If it is a _markdown file_: Filename without extension
       - If it is a _directory_: Directory name
         <a id="wiki_slug_dir"></a>

         - [ ] Start by creating a blank page, that is, create respective entries in 
           `wiki_article` and ` wiki_articlerevision`.

     - `lft`, `rght`: [Algorithm](#db_wiki_lftright_algo)

     - [level](#db_wiki_fld_level): Starting from the root upload directory on the local 
       machine
       - each _directory_ adds `1` to the level, that is, `<parent_dir_lvl> + 1`,
       - each _markdown file_ in a directory has the value `<dir_lvl> + 1`,
       - for the root directory `<dir_lvl>` and `<parent_dir_lvl>` are both `0`.

    - `parentid`: 
       - If it is a _markdown file_: `wiki_urlpath-id` of file's directory.
       - If it is a _directory_: `wiki_urlpath-id` of parent directory.


## API-Design

### GET /articles/{id} - retrieve article

  - [ ] Document API

### POST /articles - create article

#### Root article

  - [ ] Document API

#### Child articles

  In Django Wiki calling `GET <domain>/foo/bar` will redirect to the creation of a new 
  article page with slug `bar` as child of `foo`. This only works if `foo` exists.

  The API will _not support_ this behaviour.

  Instead, the API uses the `parent_art_id` provided in the JSON POST 
  payload to identify the parent node. 

  If a second child article is added under the same root node, it becomes the _right 
  sibling ("append")_ to the other existing child article. Curtently, there is no 
  specific reason why using an "append" over an "insert at the beginning".


## Installation guide

### Go project and dependencies

  1. `go mod init coco-life.de/wapi`
  1. `go get github.com/gin-gonic/gin`
  1. `go get github.com/jackc/pgx/v4`
