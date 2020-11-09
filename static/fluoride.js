// @license magnet:?xt=urn:btih:90dc5c0be029de84e523b9b3922520e79e0e6f08&dn=cc0.txt CC0

var reverseActions = {
	"like": "unlike",
	"unlike": "like",
	"retweet": "unretweet",
	"unretweet": "retweet"
};

var statusCache = {};

var csrfToken = "";
var antiDopamineMode = false;

function checkCSRFToken() {
	var tag = document.querySelector("meta[name='csrf_token']");
	if (tag)
		csrfToken = tag.getAttribute("content");
}

function checkAntiDopamineMode() {
	var tag = document.querySelector("meta[name='antidopamine_mode']");
	if (tag)
		antiDopamineMode = tag.getAttribute("content") === "true";
}

function http(method, url, body, type, success, error) {
	var req = new XMLHttpRequest();
	req.onload = function() {
		if (this.status === 200 && typeof success === "function") {
			success(this.responseText, this.responseType);
		} else if (typeof error === "function") {
			error(this.responseText);
		}
	};
	req.onerror = function() {
		if (typeof error === "function") {
			error(this);
		}
	};
	req.open(method, url);
	if (type)
		req.setRequestHeader("Content-Type", type);
	req.send(body);
}

function updateActionForm(id, f, action) {
	f.querySelector("[type='submit']").value = action;
	f.action = "/" + action + "/" + id;
	f.dataset.action = action;
}

function handleLikeForm(id, f) {
	f.onsubmit = function(event) {
		event.preventDefault();

		var action = f.dataset.action;
		var forms = document.
			querySelectorAll(".status-"+id+" .status-like");
		for (var i = 0; i < forms.length; i++) {
			updateActionForm(id, forms[i], reverseActions[action]);
		}

		var body = "csrf_token=" + encodeURIComponent(csrfToken);
		var contentType = "application/x-www-form-urlencoded";
		http("POST", "/fluoride/" + action + "/" + id, 
			body, contentType, function(res, type) {

			if (antiDopamineMode)
				return;
			var data = JSON.parse(res);
			var count = data.data;
			if (count === 0)
				count = "";
			var counts = document.
				querySelectorAll(".status-"+id+" .status-like-count");
			for (var i = 0; i < counts.length; i++) {
				if (count > 0) {
					counts[i].innerHTML = "(" + count + ")";
				} else {
					counts[i].innerHTML = "";
				}
			}
		}, function(err) {
			for (var i = 0; i < forms.length; i++) {
				updateActionForm(id, forms[i], action);
			}
		});
	}
}

function handleRetweetForm(id, f) {
	f.onsubmit = function(event) {
		event.preventDefault();

		var action = f.dataset.action;
		var forms = document.
			querySelectorAll(".status-"+id+" .status-retweet");
		for (var i = 0; i < forms.length; i++) {
			updateActionForm(id, forms[i], reverseActions[action]);
		}

		var body = "csrf_token=" + encodeURIComponent(csrfToken);
		var contentType = "application/x-www-form-urlencoded";
		http("POST", "/fluoride/" + action + "/" + id, 
			body, contentType, function(res, type) {

			if (antiDopamineMode)
				return;
			var data = JSON.parse(res);
			var count = data.data;
			if (count === 0)
				count = "";
			var counts = document.
				querySelectorAll(".status-"+id+" .status-retweet-count");
			for (var i = 0; i < counts.length; i++) {
				if (count > 0) {
					counts[i].innerHTML = "(" + count + ")";
				} else {
					counts[i].innerHTML = "";
				}
			}
		}, function(err) {
			for (var i = 0; i < forms.length; i++) {
				updateActionForm(id, forms[i], action);
			}
		});
	}
}

function isInView(el) {
	var ract = el.getBoundingClientRect();
	if (ract.top > 0 && ract.bottom < window.innerHeight)
		return true;
	return false;
}

function replyToLinkLocal(a) {
	a.onmouseenter = function(event) {
		var id = event.target.getAttribute("href");
		var status = document.querySelector(id);
		if (!status)
			return;
		if (isInView(status)) {
			status.classList.add("highlight");
		} else {
			var copy = status.cloneNode(true);
			copy.id = "reply-to-popup";
			var ract = event.target.getBoundingClientRect();
			if (ract.top > window.innerHeight / 2) {
				copy.style.bottom = (window.innerHeight - 
					window.scrollY - ract.top) + "px";
			}
			event.target.parentElement.appendChild(copy);
		}
	}
	a.onmouseleave = function(event) {
		var popup = document.getElementById("reply-to-popup");
		if (popup) {
			event.target.parentElement.removeChild(popup);    
		} else {
			var id = event.target.getAttribute("href");
			document.querySelector(id)
				.classList.remove("highlight");
		}
	}
}

var inMouseEnter = false;
function replyToLinkRemote(a) {
	a.onmouseenter = function(event) {
		inMouseEnter = true;
		var id = event.target.getAttribute("href");
		var arr = id.replace("/thread", "").split("#");
		if (arr.length < 2)
			return
		id = arr[1].replace("status-", "");
		if (statusCache[id]) {
			var copy = document.createElement("div");
			copy.innerHTML = statusCache[id];
			copy = copy.firstElementChild;
			copy.id = "reply-to-popup";
			var ract = event.target.getBoundingClientRect();
			if (ract.top > window.innerHeight / 2) {
				copy.style.bottom = (window.innerHeight - 
					window.scrollY - ract.top) + "px";
			}
			event.target.parentElement.appendChild(copy);
		} else {
			http("GET", "/fluoride/status/"+id, null, null, function(res, type) {
				statusCache[id] = res;
				if (!inMouseEnter)
					return;
				var copy = document.createElement("div");
				copy.innerHTML = res;
				copy = copy.firstElementChild;
				copy.id = "reply-to-popup";
				var ract = event.target.getBoundingClientRect();
				if (ract.top > window.innerHeight / 2) {
					copy.style.bottom = (window.innerHeight - 
						window.scrollY - ract.top) + "px";
				}
				event.target.parentElement.appendChild(copy);
			}, function(err) {
				console.log("error:", err);
			})
		}
	}
	a.onmouseleave = function(event) {
		inMouseEnter = false;
		var popup = document.getElementById("reply-to-popup");
		if (popup) {
			popup.parentElement.removeChild(popup);    
		} 
	}
}

function handleReplyToLink(a) {
	if (!a)
		return;
	var id = a.getAttribute("href");
	if (!id)
		return;
	if (id[0] === "#") {
		replyToLinkLocal(a);
	} else if (id.indexOf("/thread/") === 0) {
		replyToLinkRemote(a);
	}
}

function handleReplyLink(a) {
	a.onmouseenter = function(event) {
		var id = event.target.getAttribute("href");
		var status = document.querySelector(id);
		if (!status)
			return;
		if (isInView(status)) {
			status.classList.add("highlight");
		} else {
			var copy = status.cloneNode(true);
			copy.id = "reply-popup";
			var ract = event.target.getBoundingClientRect();
			if (ract.left > window.innerWidth / 2) {
				copy.style.right = (window.innerWidth -
					ract.right - 12) + "px";
			}
			event.target.parentElement.appendChild(copy);
		}
	}
	a.onmouseleave = function(event) {
		var popup = document.getElementById("reply-popup");
		if (popup) {
			event.target.parentElement.removeChild(popup);
		} else {
			var id = event.target.getAttribute("href");
			document.querySelector(id).classList.remove("highlight");
		}
	}
}

function handleStatusLink(a) {
	if (a.classList.contains("mention"))
		return;
	a.target = "_blank";
}

document.addEventListener("DOMContentLoaded", function() { 
	checkCSRFToken();
	checkAntiDopamineMode();

	var statuses = document.querySelectorAll(".status-container");
	for (var i = 0; i < statuses.length; i++) {
		var s = statuses[i];
		var id = s.dataset.id;

		var likeForm = s.querySelector(".status-like");
		handleLikeForm(id, likeForm);

		var retweetForm = s.querySelector(".status-retweet");
		handleRetweetForm(id, retweetForm);

		var replyToLink = s.querySelector(".status-reply-to-link");
		handleReplyToLink(replyToLink);

		var replyLinks = s.querySelectorAll(".status-reply-link");
		for (var j = 0; j < replyLinks.length; j++) {
			handleReplyLink(replyLinks[j]);
		}

		var links = s.querySelectorAll(".status-content a");
		for (var j = 0; j < links.length; j++) {
			handleStatusLink(links[j]);
		}
	}
});

// @license-end
