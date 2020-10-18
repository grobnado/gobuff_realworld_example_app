package actions

import (
	"fmt"
	"gobuff_realworld_example_app/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/pkg/errors"
)

//ArticlesRead renders the article
func ArticlesRead(c buffalo.Context) error {
	slug := c.Param("slug")

	a := []models.Article{}
	tx := c.Value("tx").(*pop.Connection)
	tx.Where("slug = ?", slug).Eager().All(&a)

	// article not found so redirect to home
	if len(a) == 0 {
		return c.Redirect(302, "/")
	}

	article := a[0]

	c.Set("article", article)

	comment := models.Comment{}
	c.Set("comment", comment)

	comments := []models.Comment{}
	tx.Where("article_id = ?", article.ID).Order("created_at desc").Limit(20).Eager().All(&comments)

	c.Set("comments", comments)

	return c.Render(200, r.HTML("articles/read.html"))
}

//ArticlesComment renders the article
func ArticlesComment(c buffalo.Context) error {
	u := c.Value("current_user").(*models.User)
	slug := c.Param("slug")

	comment := &models.Comment{}

	if err := c.Bind(comment); err != nil {
		return errors.WithStack(err)
	}

	if comment.Body == "" {
		return c.Redirect(302, fmt.Sprintf("/articles/%v", slug))
	}

	a := []models.Article{}
	tx := c.Value("tx").(*pop.Connection)
	tx.Where("slug = ?", slug).All(&a)

	// article not found so redirect to home
	if len(a) == 0 {
		return c.Redirect(302, "/")
	}

	article := a[0]

	comment.UserID = u.ID
	comment.ArticleID = article.ID

	_, err := comment.Create(tx)
	if err != nil {
		return errors.WithStack(err)
	}

	return c.Redirect(302, fmt.Sprintf("/articles/%v", slug))
}

//ArticlesNew renders the articles form
func ArticlesNew(c buffalo.Context) error {
	a := models.Article{}
	c.Set("article", a)
	return c.Render(200, r.HTML("articles/new.html"))
}

// ArticlesCreate creates a new article
func ArticlesCreate(c buffalo.Context) error {
	u := c.Value("current_user").(*models.User)

	a := &models.Article{}
	a.UserID = u.ID

	if err := c.Bind(a); err != nil {
		return errors.WithStack(err)
	}

	tx := c.Value("tx").(*pop.Connection)
	verrs, err := a.Create(tx)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("article", a)
		c.Set("errors", verrs)
		return c.Render(200, r.HTML("articles/new.html"))
	}

	c.Flash().Add("success", "Article created")

	return c.Redirect(302, fmt.Sprintf("/articles/%v", a.Slug))
}
