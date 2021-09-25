-- Insert article as topleft child under the root article.
BEGIN;

create table input (title text, content text, slug text);
insert into input values (:'title', :'content', :'slug');

DO $$
DECLARE
  wiki_article_id integer;
  wiki_articlerevision_id integer;
  title_input text;
  content_input text;
  slug_input text;
BEGIN
  select title, content, slug from input into title_input, content_input, slug_input;
  with header AS (
    insert into
      wiki_article
      (
        created,
        modified,
        group_read,
        group_write,
        other_read,
        other_write,
        current_revision_id
      )
      values
      (
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP,
        true,
        true,
        true,
        true,
        null -- revision_id has a UNIQUE constraint. We can set it once the revision is created.
      )
      returning id as hdr_id
  ),
  revision as (
    insert into
      wiki_articlerevision
      (
        article_id,
        revision_number,
        previous_revision_id,
        title,
        content,
        created,
        modified,
        deleted,
        locked,
        user_message,
        automatic_log
      )
      select 
        hdr_id,
        1,
        null,
        title_input,
        content_input,
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP,
        false,
        false,
        '',
        ''
      from header
      returning id as rev_id, article_id as fk_hdr_id
  ),
  url as (
    insert into
      wiki_urlpath
      (
        slug,
        lft,
        rght,
        level,
        tree_id,
        article_id,
        site_id,
        parent_id
      )
    select 
        slug_input,
        2,
        3,
        1,
        1,
        hdr_id,
        1,
        1
    from header
  )
  select hdr_id, rev_id into wiki_article_id, wiki_articlerevision_id
  from header, revision 
  where header.hdr_id = revision.fk_hdr_id;

  update wiki_article
    set current_revision_id = wiki_articlerevision_id
    where id = wiki_article_id;
END $$;
