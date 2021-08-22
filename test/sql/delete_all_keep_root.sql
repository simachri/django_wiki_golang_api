-- Delete all articles except for the root article.
BEGIN;

DELETE FROM wiki_urlpath WHERE level != 0;
DELETE FROM wiki_articlerevision WHERE article_id != 1;
DELETE FROM wiki_article WHERE id != 1;
