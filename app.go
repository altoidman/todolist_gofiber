package main

import (
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gofiber/storage/redis"
	"github.com/gofiber/template/html/v2"
	"github.com/jmoiron/sqlx"
)
func main() {
	storage := redis.New(redis.Config{
		Host: "127.0.0.1",
		Port: 6379,
		Database: 0,
	})
	store := session.New(session.Config{
		Storage: storage,
		KeyGenerator: utils.UUIDv4,
		KeyLookup: "cookie:_session",
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
	})
	
	app := fiber.New(fiber.Config{Views: html.New("./views",".html")})
	db,err := sqlx.Connect("mysql","root:root!@tcp(127.0.0.1:3306)/task")
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	app.Get("/",func(c *fiber.Ctx) error {
		sess,_ := store.Get(c)

		_,err := db.Exec(`
		CREATE TABLE IF NOT EXISTS lists(
		id INT AUTO_INCREMENT NOT NULL,
		username VARCHAR(50) NOT NULL UNIQUE,
		title VARCHAR(255) NOT NULL,
		text TEXT NOT NULL,
		created VARCHAR(50) NOT NULL,
		url VARCHAR(255) NOT NULL UNIQUE,
		PRIMARY KEY (id)
		);`)
		if err != nil{
			log.Fatal(err)
		}
		_,err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users(
		id INT AUTO_INCREMENT NOT NULL,
		username VARCHAR(50) NOT NULL UNIQUE,
		password TEXT NOT NULL,
		created VARCHAR(50) NOT NULL,
		PRIMARY KEY (id)
		);
		`)
		if err != nil {
			log.Fatal(err)
		}

		
		var lists [] struct {
			ID int
			Title string
			Text string
			Created string
			Url string
		}
		if sess.Get("username") == nil || sess.Get("username").(string) == ""{
			return c.Redirect("/login")
		}
		err = db.Select(&lists,"SELECT id,title,text,created,url FROM lists WHERE username = ?",sess.Get("username").(string))
		return c.Render("task",fiber.Map{"lists":lists,"username":sess.Get("username").(string)})
	})

	app.Post("/",func(c *fiber.Ctx) error {
		sess,_ := store.Get(c)
		username,ok := sess.Get("username").(string)
		if !ok || username == "" {
			return c.Redirect("/login")
		}
		title := c.FormValue("title")
		content := c.FormValue("content")
		created := time.Now().Format("Jan 2, 2006 15:04:05")
		if title == "" && content == ""{
			return c.Render("task",fiber.Map{"err":"please check data is empty" ,"username":sess.Get("username").(string)})
		}
		rand := utils.UUIDv4()+utils.UUIDv4()+utils.UUIDv4()
		_,err = db.Exec("INSERT INTO lists (username,title,text,created,url) VALUES(?,?,?,?,?)",username,title,content,created,rand)

		if err != nil{
			return c.Render("task",fiber.Map{"err":"error for save data????","username":sess.Get("username").(string)})
		}
		c.Render("task",fiber.Map{"success":"successfully Create new list/task","username":sess.Get("username").(string)})
		time.Sleep(500 * time.Millisecond)
		return c.Redirect("/")
		
	})

	app.Get("/login",func(c *fiber.Ctx) error {
		return c.Render("login",fiber.Map{})
	})
	app.Post("/login",func(c *fiber.Ctx) error {
		sess,err := store.Get(c);
		if err != nil {
			log.Fatal(err)
		}
		username := c.FormValue("username")
		password := c.FormValue("password")
		if username == "" || password == "" {
			return c.Render("login",fiber.Map{"msg":"username or password not right"})
		}
		var user struct {
			Username string
			Password string
		}
		err = db.Get(&user,"SELECT username,password FROM users WHERE username = ?",username)
		if err != nil && user.Password != password{
			return c.Render("login",fiber.Map{"msg":"username or password not right"})
		}
		sess.Set("username",username)
		sess.SetExpiry(5 * time.Hour)
		if err:=sess.Save(); err != nil{
			log.Fatal(err)
		}
		return c.Render("login",fiber.Map{"success":"Successfully login now ok(())"})
	})

	app.Get("/DELETE",func(c *fiber.Ctx) error {
		url := c.Query("delete")
		_,err := db.Exec("DELETE FROM lists WHERE url = ?",url)
		if err != nil {
			return c.Render("task",fiber.Map{"err":"Error For DELETE List ?????"})
		}
	return c.Redirect("/")
	})

	app.Get("logout",func(c *fiber.Ctx) error {
		sess, _ := store.Get(c)
		sess.Destroy()
		return c.Redirect("/login");
	})

app.Listen(":8080")

}
