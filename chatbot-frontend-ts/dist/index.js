"use strict";
document.getElementById("send-button").addEventListener("click", function () {
    const userQuery = document.getElementById("user-query").value;
    if (userQuery) {
        sendMessage(userQuery);
    }
});
// Function to send a message to the Go backend
function sendMessage(query) {
    const chatBox = document.getElementById("chat-box");
    chatBox.innerHTML += `<div>User: ${query}</div>`; // Display user message
    fetch("http://127.0.0.1:8080/query", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({ query: query }) // Send the query to the Go backend
    })
        .then((response) => response.json())
        .then((data) => {
            chatBox.innerHTML += `<div>Bot: ${data.response}</div>`; // Display bot response
        })
        .catch((error) => {
            chatBox.innerHTML += `<div>Error: ${error.message}</div>`; // Display any errors
        });
}