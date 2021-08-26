# Django-Wiki API in Golang

  The API uses a Golang API server [Gin](https://github.com/gin-gonic/gin) and directly 
  updates the Postgres database through the Golang driver 
  [pgx](https://github.com/jackc/pgx). There is an alternative way to add API functions 
  to Django itself and call them from Golang API server instead of directly updating the 
  database.  The decision to directly interact with the database is on purpose to be 
  independent from the Django implementation and learn something about working 
  programmatically with a Postgres database.


## To do

  - [X] Golang API: DB-Zugriff
  - [ ] RESTful-API-Design
    - [ ] Generell einlernen
    - [ ] Ist es sinnvoll, SWAGGER o.ä. einzusetzen? Wenn ja, wie funktioniert es?
  - [ ] Implementierung


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

### Go tests

  From the project root run `go test -v ./...`.

### Interactively

  1. Spin up the Docker containers:
     ```sh
     $ cd ~/Development/Linode-coco-life/django-on-docker
     $ docker-compose down
     $ docker-compose up -d
     $ docker-compose logs
     ```

  1. To connect to the PostgreSQL database from outside the Docker container use:
     1. Take the password from 
        `~/Development/Linode-coco-life/django-on-docker/.env.dev.db`
     1. `psql -U xi3k -h 0.0.0.0 -d pk_db_dev`

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

### Database: Fill with test data

  Use `psql`:
  - `\set title 'New article'`
  - `\set content '# First level header'`
  - `\set slug 'foo'`
  - `\i test/sql/new_article.sql`


### Database: Clear all test data

  Use `psql`: `\i test/sql/delete_all_keep_root.sql`


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
    - Is _empty_ for the __root node__.
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
      - _Insert_ a new article `n` as _sibling_ next to node `l` (now on the left) as a 
        _leaf_ node under _parent_ `p`:
        1. Set `lft` and `rght` of `n` based on `l`:
           - `n.lft = l.rght + 1`
           - `n.rght = l.rght + 2`
        1. Adjust all nodes `r` right to `n` that are under _parent_ `p`, that is, all 
           nodes `r` with `n.rght >= r.lft AND r.rght < p.rght`:
           - Set `r.lft = r.lft + 2`
           - Set `r.rght = r.rght + 2`
         1. Adjust all nodes `r` above or right to _parent_ `p` including `p` istelf:
            <a id="adj_nodes_above_and_right"></a>

            - `rght`: All nodes `r` with `n.rght <= r.rght` (direct parent and 
              ancestors): Set `r.rght = r.rght + 2`
            - `lft`: All nodes `r` with `n.rght <= r.lft` (nodes right to direct parent 
              or ancestors): Set `r.lft = r.lft + 2`

      - _Insert_ a new article `n` as _child_ to _parent_ `p`.
        1. Set `lft` and `rght` of `n` based on `p`:
           - `n.lft = p.lft + 1`
           - `n.rght = p.lft + 2`
         1. Adjust all nodes `r` above or right to _parent_ `p` including `p` istelf.
            Same [algorithm](#adj_nodes_above_and_right).

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

### /articles


## Installation guide

### Go project and dependencies

  1. `go mod init coco-life.de/wapi`
  1. `go get github.com/gin-gonic/gin`
  1. `go get github.com/jackc/pgx/v4`
