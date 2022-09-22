let posts = document.getElementById("post-feed");
let onlineUsers = document.getElementById("onlineusers");

let postButton = document.getElementById("new-post-btn");

let users = ["tb38r", "abmutungi", "eternal17", "million"];

for (let i = 0; i < 10; i++) {
  let postDivs = document.createElement("div");
  let postTitle = document.createElement("div");
  postTitle.id = i;
  postTitle.className = "post-title-class";
  let postContent = document.createElement("div");
  postContent.id = i;
  postContent.className = "post-content-class";

  let postFooter = document.createElement("div");
  postFooter.id = i;
  postFooter.className = "post-footer-class";
  postDivs.className = "post-class ";
  postDivs.id = i;
  postTitle.innerText = `This is post number ${i}\n`;
  postContent.innerText =
    " This is a post bla blablalala\n___________________________________________________";
  postFooter.innerText = `Created by abmutungi,   Date: ${new Date().toDateString()}, Comments: ${i + 13}`;
  postDivs.appendChild(postTitle);
  postDivs.appendChild(postContent);
  postDivs.appendChild(postFooter);

  posts.appendChild(postDivs);
}

let userDetails;
let imageDiv;
let img;

for (let i = 0; i < 4; i++) {
  userDetails = document.createElement("div");
  let username = document.createElement("div");
  imageDiv = document.createElement("div");
  img = document.createElement("img");
  let onlineIcon = document.createElement("div");

  onlineIcon.className = "online-icon-class";

  img.src = "/css/img/newcastle.png";
  img.style.width = "2vw";
  imageDiv.appendChild(onlineIcon);
  userDetails.id = `${users[i]}`;

  //   userDetails.setAttribute("type", "button");

  userDetails.className = "registered-user";
  username.innerText = `${users[i]}`;
  imageDiv.append(img);
  userDetails.appendChild(username);
  userDetails.appendChild(imageDiv);
  onlineUsers.appendChild(userDetails);
}

let postTitlesClick = document.getElementsByClassName("post-title-class");
Array.from(postTitlesClick).forEach(function (postTitle) {
  postTitle.addEventListener("click", function (e) {
    displayPostModal.style.display = "block";
  });
});

let modal = document.getElementsByClassName("modal");
let chatModal = document.getElementById("my-chat-modal");
let createPostModal = document.getElementById("create-post-modal");
let displayPostModal = document.getElementById("display-post-modal");

postButton.addEventListener("click", function () {
  createPostModal.style.display = "block";
});

let userRg = document.querySelectorAll(".registered-user");
let chatRecipient = document.getElementById("chat-recipient");

// Get the button that opens the modal
let btn = document.getElementById("myBtn");

// Get the <span> element that closes the modal
let span = document.getElementsByClassName("close");

// When the user clicks the button, open the modal

for (let i = 0; i < userRg.length; i++) {
  userRg[i].onclick = function () {
    chatRecipient.innerText = userRg[i].id;

    console.log("Users clicked");
    chatModal.style.display = "block";
  };
}

// When the user clicks on <span> (x), close the modal
for (let i = 0; i < span.length; i++) {
  span[i].onclick = function () {
    modal[i].style.display = "none";
  };
}

// When the user clicks anywhere outside of the modal, close it
window.onclick = function (event) {
  for (let i = 0; i < modal.length; i++) {
    // console.log("modal -> ", modal[i]);
    // console.log("evt -> ", event.target);
    if (event.target == modal[i]) {
      modal[i].style.display = "none";
    }
  }
};

let sendArrow = document.getElementById("chat-arrow");
let chatTextArea = document.getElementById("chat-input");
let chatContainer = document.getElementById("chat-container");
let chatBody = document.getElementById("chat-box-body");
let displayPostBody = document.getElementById("display-post-body");
let sender = true;

sendArrow.addEventListener("click", function () {
  console.log("arrow clicked");
  let newChatBubble = document.createElement("div");

  newChatBubble.innerText = chatTextArea.value;
  chatTextArea.value = "";
  if (sender) {
    newChatBubble.id = "chat-message-sender";
    sender = false;
  } else {
    newChatBubble.id = "chat-message-recipient";
    sender = true;
  }

  chatContainer.appendChild(newChatBubble);

  chatBody.scrollTo(0, chatBody.scrollHeight);
});



const teamCrests = [
  "/css/img/newcastle.png",
  "/css/img/chelsea.png",
  "/css/img/man-u.png",
  "/css/img/man-city.png",
  "/css/img/liverpool.png",
  "/css/img/spurs.png",
];

const categorySelection = document.getElementById("category-selection");



for (let i = 0; i < teamCrests.length; i++) {
  let img = document.createElement("img");
  img.style.backgroundColor = 'white'
  img.alt = "none"
  img.id = teamCrests[i].slice(teamCrests[i].lastIndexOf("/") + 1, teamCrests[i].length - 4);
  img.classList = "crest-colors";
  img.src = teamCrests[i];
  categorySelection.append(img);
}

let crestcolors = document.getElementsByClassName("crest-colors");

const colorSwitch = {
    newcastle : `linear-gradient(
      to right,
      #040108,
      #040108 50%,
      #f0f0f0 50%,
      #f0f0f0
    )`,
    spurs : "lightgrey", 
    "man-u" : "red",
    chelsea: "blue",
    liverpool : "red",
    "man-city": "skyblue"

}; 

for (let i = 0; i < crestcolors.length; i++) {
  crestcolors[i].addEventListener("mouseup", (e) => {
  
    if( e.target.alt == 'none'){
        e.target.style.background = colorSwitch[e.target.id]
        e.target.alt = colorSwitch[e.target.id]

    }else{
        e.target.style.background = "white"
        e.target.alt = "none"

    }
  

  });
}
let commentContainer = document.getElementById("comment-container");
let commentArrow = document.getElementById("comment-arrow");
let commentTextArea = document.getElementById("comment-input");

commentArrow.addEventListener("click", function () {
  let i = 0;
  let comment = document.createElement("div");
  let commentDetails = document.createElement("div");
  commentDetails.innerText = `Created by: McTom Date: ${
    new Date().toISOString().split("T")[0]
  } ${new Date().toISOString().split("T")[1].substring(0, 5)}`;
  comment.style.marginBottom = "1vh";
  comment.id = `comment-${i}`;
  commentDetails.id = `comment-detail-${i}`;
  comment.innerText = `${commentTextArea.value}`;
  commentTextArea.value = "";
  comment.appendChild(commentDetails);
  commentContainer.appendChild(comment);
  displayPostBody.scrollTo(0, displayPostBody.scrollHeight);
});