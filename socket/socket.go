package socket

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"real-time-forum/chat"
	"real-time-forum/comments"
	notification "real-time-forum/notifications"
	"real-time-forum/posts"
	"real-time-forum/users"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

type T struct {
	TypeChecker
	*posts.Posts
	*comments.Comments
	*Register
	*Login
	*Logout
	*comments.CommentsFromPosts
	*chat.Chat
	*whosNotifications
	*deleteNotifications
	*TypingNotification
	*TypingNotificationEnd
}

type TypeChecker struct {
	Type string `json:"type"`
}

// type Posts struct {
// 	Title       string `json:"title"`
// 	PostContent string `json:"postcontent"`
// 	Date        string `json:"posttime"`
// 	Tipo        string `json:"tipo"`
// 	Username    string `json:"username"`
// }

type Register struct {
	Username  string `json:"username"`
	Age       string `json:"age"`
	Email     string `json:"email"`
	Gender    string `json:"gender"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Password  string `json:"password"`
	Team      string `json:"teamreg"`
	Tipo      string `json:"tipo"`
}

type Login struct {
	LoginUsername string `json:"loginUsername"`
	LoginPassword string `json:"loginPassword"`
	Tipo          string `json:"tipo"`
}

type Logout struct {
	LogoutUsername string `json:"logoutUsername"`
	Tipo           string `json:"tipo"`
	LogoutClicked  string `json:"logoutClicked"`
}

type formValidation struct {
	UsernameLength         bool             `json:"usernameLength"`
	UsernameSpace          bool             `json:"usernameSpace"`
	UsernameDuplicate      bool             `json:"usernameDuplicate"`
	EmailDuplicate         bool             `json:"emailDuplicate"`
	PasswordLength         bool             `json:"passwordLength"`
	AgeEmpty               bool             `json:"ageEmpty"`
	FirstNameEmpty         bool             `json:"firstnameEmpty"`
	LastNameEmpty          bool             `json:"lastnameEmpty"`
	EmailInvalid           bool             `json:"emailInvalid"`
	SuccessfulRegistration bool             `json:"successfulRegistration"`
	AllUserAfterNewReg     []users.AllUsers `json:"allUserAfterNewReg"`
	OnlineUsers            []string         `json:"onlineUsers"`
	Tipo                   string           `json:"tipo"`
}
type loginValidation struct {
	InvalidUsername    bool             `json:"invalidUsername"`
	InvalidPassword    bool             `json:"invalidPassword"`
	SuccessfulLogin    bool             `json:"successfulLogin"`
	SuccessfulUsername string           `json:"successfulusername"`
	Tipo               string           `json:"tipo"`
	SentPosts          []posts.Posts    `json:"dbposts"`
	AllUsers           []users.AllUsers `json:"allUsers"`
	OnlineUsers        []string         `json:"onlineUsers"`
	UsersWithChat      []chat.Chat      `json:"userswithchat"`
	PopUserCheck       string           `json:"popusercheck"`
}

type whosNotifications struct {
	Username string
}

type notificationsAtLogin struct {
	Notifications []notification.Notification `json:"notification"`
	Response      string                      `json:"response"`
	UserToDelete  string                      `json:"usertodelete"`
	Tipo          string                      `json:"tipo"`
}

type deleteNotifications struct {
	NotificationSender    string `json:"sender"`
	NotificationRecipient string `json:"recipient"`
}

type updateOnlineUsers struct {
	UpdatedOnlineUsers []string         `json:"UpdatedOnlineUsers"`
	Tipo               string           `json:"tipo"`
	AllUsers           []users.AllUsers `json:"updateAllUsers"`
}

type TypingNotification struct {
	TypingRecipient string `json:"typingrecipient"`
	TypingSender    string `json:"typingsender"`
}

type SendTypingNotification struct {
	TypingRecipient string `json:"typingRecipient"`
	TypingSender    string `json:"typingSender"`
	Tipo            string `json:"tipo"`
}

type TypingNotificationEnd struct {
	TypingEndRecipient string `json:"typingrecipient"`
	TypingEndSender    string `json:"typingsender"`
}

type SendTypingNotificationEnd struct {
	TypingEndRecipient string `json:"typingEndRecipient"`
	TypingEndSender    string `json:"typingEndSender"`
	Tipo               string `json:"tipo"`
}

var (
	loggedInUsers         = make(map[string]*websocket.Conn)
	broadcastChannelPosts = make(chan posts.Posts, 1)

	broadcastChannelComments = make(chan comments.Comments, 1)
	currentUser              = ""
	CallWS                   = false
	online                   loginValidation
	broadcastOnlineUsers     = make(chan updateOnlineUsers, 1)
	notifyAtLogin            notificationsAtLogin
)

// unmarshall data based on type
func (t *T) UnmarshalForumData(data []byte) error {
	if err := json.Unmarshal(data, &t.TypeChecker); err != nil {
		log.Println("Error when trying to sort forum data type...")
	}

	switch t.Type {
	case "post":
		t.Posts = &posts.Posts{}
		return json.Unmarshal(data, t.Posts)
	case "comment":
		t.Comments = &comments.Comments{}
		return json.Unmarshal(data, t.Comments)
	case "signup":
		t.Register = &Register{}
		return json.Unmarshal(data, t.Register)
	case "login":
		t.Login = &Login{}
		return json.Unmarshal(data, t.Login)
	case "logout":
		t.Logout = &Logout{}
		return json.Unmarshal(data, t.Logout)
	case "getcommentsfrompost":
		t.CommentsFromPosts = &comments.CommentsFromPosts{}
		return json.Unmarshal(data, t.CommentsFromPosts)
	case "chatMessage":
		t.Chat = &chat.Chat{}
		return json.Unmarshal(data, t.Chat)

	case "requestChatHistory":
		t.Chat = &chat.Chat{}
		return json.Unmarshal(data, t.Chat)

	case "requestNotifications":
		t.whosNotifications = &whosNotifications{}
		return json.Unmarshal(data, t.whosNotifications)
	case "deletenotification":
		t.deleteNotifications = &deleteNotifications{}
		return json.Unmarshal(data, t.deleteNotifications)
	case "typingnotificationstart":
		t.TypingNotification = &TypingNotification{}
		return json.Unmarshal(data, t.TypingNotification)
	case "typingnotificationend":
		t.TypingNotificationEnd = &TypingNotificationEnd{}
		return json.Unmarshal(data, t.TypingNotificationEnd)
	default:
		return fmt.Errorf("unrecognized type value %q", t.Type)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func WebSocketEndpoint(w http.ResponseWriter, r *http.Request) {
	db, _ := sql.Open("sqlite3", "real-time-forum.db")

	go broadcastToAllClients()
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("error when upgrading connection...")
	}
	fmt.Println("CONNECTION TO CLIENT")
	defer wsConn.Close()

	// if user logs into 2 clients, close first connection
	if _, ok := loggedInUsers[currentUser]; ok {
		loggedInUsers[currentUser].Close()
	}

	loggedInUsers[users.GetUserName(db, currentUser)] = wsConn

	fmt.Println("LOGGED IN USERS", loggedInUsers)

	online.Tipo = "onlineUsers"

	online.OnlineUsers = []string{}

	var o updateOnlineUsers
	for k := range loggedInUsers {

		online.UsersWithChat = chat.GetLatestChat(db, chat.GetChat(db, k))
		online.PopUserCheck = currentUser
		online.OnlineUsers = append(online.OnlineUsers, k)
		online.AllUsers = users.GetAllUsers(db)
		loggedInUsers[k].WriteJSON(online)

		o.UpdatedOnlineUsers = append(o.UpdatedOnlineUsers, k)
	}

	o.AllUsers = users.GetAllUsers(db)
	o.Tipo = "updatedOnlineUsers"

	broadcastOnlineUsers <- o

	var f T
	for {
		message, info, _ := wsConn.ReadMessage()
		fmt.Println("----", string(info))

		// if a connection is closed, we return out of this loop
		if message == -1 {
			fmt.Println("connection closed")
			for username, socketConnection := range loggedInUsers {
				if wsConn == socketConnection {
					delete(loggedInUsers, username)
				}
			}
			fmt.Println("users left in array", loggedInUsers)
			online.OnlineUsers = []string{}
			o.UpdatedOnlineUsers = []string{}
			online.Tipo = "loggedOutUser"

			for k := range loggedInUsers {
				online.OnlineUsers = append(online.OnlineUsers, k)
				o.UpdatedOnlineUsers = append(o.UpdatedOnlineUsers, k)
			}

			broadcastOnlineUsers <- o

			// wsConn.WriteJSON(online)
			return
		}
		f.UnmarshalForumData(info)

		if f.Type == "post" {
			f.Posts.Tipo = "post"

			posts.StorePosts(db, f.Posts.Username, f.Posts.PostTitle, f.Posts.PostContent, f.Posts.Categories)
			// posts.GetCommentData(db, 1)
			fmt.Println("this is the post title", f.PostContent)

			// STORE POSTS IN DATABASE
			broadcastChannelPosts <- posts.SendLastPostInDatabase(db)
		} else if f.Type == "comment" {

			// STORE COMMENTS IN THE DATABSE
			postID, _ := strconv.Atoi(f.Comments.PostID)
			comments.StoreComment(db, f.Comments.User, postID, f.Comments.CommentContent)

			f.Comments.Tipo = "comment"
			// wsConn.WriteJSON(comments.GetLastComment(db))
			broadcastChannelComments <- comments.GetLastComment(db)

			// broadcastChannelComments <- f.Comments
		} else if f.Type == "getcommentsfrompost" {
			// Display all comments in a post to a single user.

			var commentsfromPost posts.Posts
			clickedPostID, _ := strconv.Atoi(f.CommentsFromPosts.ClickedPostID)
			commentsfromPost.Tipo = "allComments"
			commentsfromPost.Comments = comments.DisplayAllComments(db, clickedPostID)

			fmt.Println("comments from post struct when unmarshalled", f.CommentsFromPosts)
			f.CommentsFromPosts.Tipo = "commentsfrompost"
			fmt.Println("all comments in this post", comments.DisplayAllComments(db, clickedPostID))
			wsConn.WriteJSON(commentsfromPost)
		} else if f.Type == "logout" {
			f.Logout.LogoutClicked = "true"
			fmt.Println("LOGOUT USERNAME", f.Logout.LogoutUsername)
			wsConn.WriteJSON(f.Logout)
		} else if f.Type == "chatMessage" {
			if !chat.ChatHistoryValidation(db, f.Chat.ChatSender, f.Chat.ChatRecipient).Exists {
				chat.StoreChat(db, f.Chat.ChatSender, f.Chat.ChatRecipient)
			}
			fmt.Println("THIS IS THE CHAT ID", chat.ChatHistoryValidation(db, f.Chat.ChatSender, f.Chat.ChatRecipient).ChatID)
			// then store messages using chat id
			chat.StoreMessages(db, chat.ChatHistoryValidation(db, f.Chat.ChatSender, f.Chat.ChatRecipient).ChatID, f.Chat.ChatMessage, f.Chat.ChatSender, f.Chat.ChatRecipient)

			if !notification.CheckNotification(db, f.Chat.ChatSender, f.Chat.ChatRecipient) {
				notification.AddFirstNotificationForUser(db, f.Chat.ChatSender, f.Chat.ChatRecipient)

				// fmt.Println("*********1st****************WHO IS THIS??????******", notification.NotificationQuery(db, f.Chat.ChatRecipient))
			} else {
				notification.IncrementNotifications(db, f.Chat.ChatSender, f.Chat.ChatRecipient)
			}

			fmt.Println("THIS IS CHAT HISTORY --> ", chat.GetAllMessageHistoryFromChat(db, chat.ChatHistoryValidation(db, f.Chat.ChatSender, f.Chat.ChatRecipient).ChatID))
			fmt.Println("From JS-->", f.Chat.ChatMessage, f.Chat.ChatSender)
			for user, connection := range loggedInUsers {
				if user == f.Chat.ChatSender || user == f.Chat.ChatRecipient {
					f.Chat.Tipo = "lastMessage"
					// f.Chat.Tipo = "lastnotification"

					f.Chat.LastNotification = notification.SingleNotification(db, f.Chat.ChatSender, f.Chat.ChatRecipient)
					f.Chat.UsersWithChat = chat.GetLatestChat(db, chat.GetChat(db, currentUser))
					f.Chat.AllUsers = users.GetAllUsers(db)
					connection.WriteJSON(f.Chat)

					fmt.Println("single notification test =========>", notification.SingleNotification(db, f.Chat.ChatSender, f.Chat.ChatRecipient))

				}
				//  check how many live notifications there are and send to recipient
			}

		} else if f.Type == "requestChatHistory" {
			if notification.RemoveNotifications(db, f.Chat.ChatRecipient, f.Chat.ChatSender) {
				notifyAtLogin.Response = "Notification viewed and set to nil"
				notifyAtLogin.UserToDelete = f.Chat.ChatRecipient
				loggedInUsers[f.Chat.ChatSender].WriteJSON(notifyAtLogin)
			}

			fmt.Println("sender and recipient-------", f.Chat.ChatSender, f.Chat.ChatRecipient)
			if chat.ChatHistoryValidation(db, f.Chat.ChatSender, f.Chat.ChatRecipient).Exists {
				fmt.Println("THIS IS CHAT HISTORY --> ", chat.GetAllMessageHistoryFromChat(db, chat.ChatHistoryValidation(db, f.Chat.ChatSender, f.Chat.ChatRecipient).ChatID))
				wsConn.WriteJSON(chat.GetAllMessageHistoryFromChat(db, chat.ChatHistoryValidation(db, f.Chat.ChatSender, f.Chat.ChatRecipient).ChatID))
			} else if !chat.ChatHistoryValidation(db, f.Chat.ChatSender, f.Chat.ChatRecipient).Exists {

				wsConn.WriteJSON(chat.GetAllMessageHistoryFromChat(db, chat.ChatHistoryValidation(db, f.Chat.ChatSender, f.Chat.ChatRecipient).ChatID))

			}
		} else if f.Type == "requestNotifications" {

			data := notification.NotificationQuery(db, f.whosNotifications.Username)
			// fmt.Println("notification ------- Data", data)
			notifyAtLogin.Notifications = []notification.Notification{}
			for _, value := range data {
				if f.whosNotifications.Username == value.NotificationRecipient {
					// value.Tipo = "clientnotifications"
					notifyAtLogin.Notifications = append(notifyAtLogin.Notifications, value)
					notifyAtLogin.Tipo = "clientnotifications"

				}
			}
			// fmt.Println("notes---", notifyAtLogin.Notifications)
			loggedInUsers[f.whosNotifications.Username].WriteJSON(notifyAtLogin)

		} else if f.Type == "deletenotification" {
			// fmt.Println("DELETE ALL NOTIFICATIONS")
			notification.RemoveNotifications(db, f.deleteNotifications.NotificationSender, f.deleteNotifications.NotificationRecipient)
		} else if f.Type == "typingnotificationstart" {
			fmt.Println("Checking if typing notification is getting sent ---> ", f.TypingNotification)
			var t SendTypingNotification
			t.Tipo = "typingNotification"
			t.TypingRecipient = f.TypingRecipient
			t.TypingSender = f.TypingSender
			for user, conn := range loggedInUsers {
				if user == f.TypingNotification.TypingRecipient {
					conn.WriteJSON(t)
				}
			}
		} else if f.Type == "typingnotificationend" {
			var t SendTypingNotificationEnd
			t.Tipo = "typingIsOver"
			t.TypingEndRecipient = f.TypingEndRecipient
			t.TypingEndSender = f.TypingEndSender
			for user, conn := range loggedInUsers {
				if user == f.TypingNotificationEnd.TypingEndRecipient {
					conn.WriteJSON(t)
				}
			}

		}

		// log.Println("Checking what's in f ---> ", f.Chat)
	}
}

func broadcastToAllClients() {
	for {
		select {
		case post, ok := <-broadcastChannelPosts:
			if ok {
				for _, user := range loggedInUsers {

					user.WriteJSON(post)
					// fmt.Printf("Value %v was read.\n", post)
				}
			}
		case comment, ok := <-broadcastChannelComments:
			if ok {
				for _, user := range loggedInUsers {
					user.WriteJSON(comment)
					fmt.Println("LINE 248", comment)
				}
			}

		case onlineuser, ok := <-broadcastOnlineUsers:

			if ok {
				for _, user := range loggedInUsers {
					user.WriteJSON(onlineuser)
				}
			}

			// POTENTIAL WAY TO SEND CHAT TO SPECIFIC USERS
			// case onlineuser, ok := <-broadcastOnlineUsers:

			// if ok {
			// 	for userName, userConn := range loggedInUsers {
			// 		var chat chat.Chat
			// 		if userName == chat.ChatSender || userName == chat.ChatRecipient{

			// 			userConn.WriteJSON(onlineuser)
			// 		}
			// 	}
			// }

			// BROADCAST TO EVERYONE WITH A WEBSOCKET ALL ONLINE USERS

		}
	}
}

func GetLoginData(w http.ResponseWriter, r *http.Request) {
	db, _ := sql.Open("sqlite3", "real-time-forum.db")
	fmt.Println(r.Method)

	var t T

	data, _ := io.ReadAll(r.Body)

	t.UnmarshalForumData(data)

	if t.Type == "signup" {

		var u formValidation
		u.Tipo = "formValidation"
		canRegister := true

		if len(t.Register.Username) < 5 {
			u.UsernameLength = true
			canRegister = false
		}

		intAge, _ := strconv.Atoi(t.Register.Age)
		if intAge < 16 {
			fmt.Println(t.Register.Age)
			fmt.Println("age invalid")
			u.AgeEmpty = true
			canRegister = false
		}
		if t.Register.FirstName == "" {
			fmt.Println("first name empty")
			u.FirstNameEmpty = true
			canRegister = false
		}
		if t.Register.LastName == "" {
			fmt.Println("last name empty")
			u.LastNameEmpty = true
			canRegister = false
		}

		if len(t.Register.Password) < 5 {
			u.PasswordLength = true
			canRegister = false
		}

		if strings.Contains(t.Register.Username, " ") {
			u.UsernameSpace = true
			canRegister = false
		}

		if len(t.Register.Password) < 5 {
			u.PasswordLength = true
			canRegister = false
		}

		if !users.ValidEmail(t.Register.Email) {
			u.EmailInvalid = true
			canRegister = false
		}
		if users.UserExists(db, t.Register.Username) {
			u.UsernameDuplicate = true
			canRegister = false
		}

		if users.EmailExists(db, t.Register.Email) {
			u.EmailDuplicate = true
			canRegister = false
		}

		// all validations passed
		if canRegister {
			// hash password
			var hash []byte
			hash, err := bcrypt.GenerateFromPassword([]byte(t.Password), bcrypt.DefaultCost)
			if err != nil {
				fmt.Println("bcrypt err:", err)
			}
			users.RegisterUser(db, t.Register.Username, t.Register.Age, t.Gender, t.FirstName, t.LastName, hash, t.Email, t.Team)

			// data gets marshalled and sent to client
			u.SuccessfulRegistration = true
			u.AllUserAfterNewReg = users.GetAllUsers(db)
			toSend, _ := json.Marshal(u)
			fmt.Println("toSend -- > ", toSend)
			w.Write(toSend)
			// http.HandleFunc("/ws", WebSocketEndpoint)
		} else {

			toSend, _ := json.Marshal(u)
			w.Write(toSend)
		}
	}

	if t.Type == "login" {
		// validate values then
		var loginData loginValidation

		loginData.Tipo = "loginValidation"

		if !users.UserExists(db, t.Login.LoginUsername) && !users.EmailExists(db, t.Login.LoginUsername) {
			fmt.Println("Checking f.login.loginusername --> ", t.Login.LoginUsername)
			loginData.InvalidUsername = true
			toSend, _ := json.Marshal(loginData)
			w.Write(toSend)

		} else if users.UserExists(db, t.Login.LoginUsername) || users.EmailExists(db, t.Login.LoginUsername) {
			fmt.Println("user exists")
			if !users.CorrectPassword(db, t.Login.LoginUsername, t.Login.LoginPassword) {
				loginData.InvalidPassword = true
				toSend, _ := json.Marshal(loginData)
				w.Write(toSend)

			} else {

				currentUser = t.Login.LoginUsername
				loginData.SentPosts = posts.SendPostsInDatabase(db)
				loginData.AllUsers = users.GetAllUsers(db)
				loginData.UsersWithChat = chat.GetLatestChat(db, chat.GetChat(db, users.GetUserName(db, currentUser)))
				// loginData.Notifications = notification.NotificationQuery(db, t.Login.LoginUsername)
				loginData.SuccessfulLogin = true
				loginData.SuccessfulUsername = users.GetUserName(db, currentUser)
				toSend, _ := json.Marshal(loginData)

				w.Write(toSend)

				// this function upgrades the connection to a websocket.

				// go http.HandleFunc("/ws", WebSocketEndpoint)

				fmt.Println("SUCCESSFUL LOGIN")
			}

			// Check username exists
			// Check the password matches
		}

		// data gets marshalled and sent to client

	}
}
