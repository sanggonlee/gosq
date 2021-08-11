
ListStories
```sql
SELECT
  "story".*
  {{
    [if] IncludePages [then]
    ,COALESCE(
      json_agg(
        json_build_object(
          "id", "page"."id",
          "pageIndex", "page"."pageIndex",
          "title", "page"."title",
          ...
        ) ORDER BY "page"."pageIndex"
      ) FILTER (WHERE "page".id IS NOT NULL), '[]'
    ) AS pages
  }}
  ,(
    CASE WHEN "story"."version" = 1 THEN
      json_build_object(
        'uuid', "posterData"->'uuid',
        'type', "posterData"->'type',
        ...
      )
    ELSE
      json_build_object(
        'uuid', "visual"."uuid",
        "publisher_slug", "visual"."publisher_slug",
        ...
      )
    END
  ) AS "posterData"
FROM "story"
LEFT JOIN visual AS poster_visual ON poster_visual.uuid = story.poster_visual_uuid
{{
  [if] IncludePages [then]
  LEFT JOIN page ON "page"."storyId" = "story"."id"
  LEFT JOIN visual ON "visual"."uuid" = "page"."visual_uuid"
}}
WHERE "deletedAt" IS NULL
{{ [if] PublisherSlug != "" [then] (AND "story"."publisherSlug" = PublisherSlug) }}
{{ [if] len(Statuses) > 0 [then] (AND map(Statuses s => ("story"."status" = s).join(OR))) }}
{{ [if] StoryType != "" [then] (AND "story"."storyType" = StoryType) }}
{{ [if] SearchQuery != "" [then] (AND to_tsvector("story"."title") @@ plainto_tsquery( SearchQuery )) }}
GROUP BY
  story.id
  ,poster_visual.uuid
  {{ [if] IncludePages [then] ,story.id }}
OFFSET {{ PageOffset * PageSize }}
LIMIT {{ PageSize }}
ORDER BY
{{
  [if] SortBy == "title" [then] "title"
  [if] SortBy == "status" [then] "status"
  [if] SortBy == "random" [then] RANDOM()
  [else] "updatedAt"
}} {{
  [if] SortDirection == "desc" [then] DESC
  [else] ASC
}}
```