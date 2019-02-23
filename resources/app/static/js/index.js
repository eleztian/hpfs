var flag = false
var config;

let index = {
    about: function (html) {
        let c = document.createElement("div");
        c.innerHTML = html;
        asticode.modaler.setContent(c);
        asticode.modaler.show();
    },
    init: function () {
        // Init
        asticode.loader.init();
        asticode.modaler.init();
        asticode.notifier.init();

        // Wait for astilectron to be ready
        document.addEventListener('astilectron-ready', function () {
            // Listen
            index.listen();
            index.login();
            index.register();
            index.exploreSharedDirs();
            index.searchFiles();
            index.showGitLog();
            index.searchFiles();
            index.mount();
        })
    },
    showGitLog: function (name) {
        if (typeof name === "undefined") {
            return
        }
        let message = { "name": "showGitLog", payload: name };
        console.log("sdsd", name);
        // Send message
        asticode.loader.show();
        astilectron.sendMessage(message, function (message) {
            asticode.loader.hide();
            // Check error
            if (message.name === "error") {
                asticode.notifier.error(message.payload);
                return
            }
            console.log(message.payload);
            let idoc = document.getElementById("logInfo").contentWindow.document;

            let infoEle = idoc.getElementById("info");
            infoEle.innerHTML = "";
            let gitTable = idoc.createElement("table");
            gitTable.className = "gitInfoTable  table table-condensed";
            gitTable.id = "gitInfoTable";
            infoEle.appendChild(gitTable);

            let tbody = idoc.createElement("tbody");

            for (let i = 0; i < message.payload.length; i++) {
                let info = message.payload[i];
                let lenFile = 0
                if (info.FileStat !== null) {
                    lenFile = info.FileStat.length;
                }

                let tr = idoc.createElement("tr");
                tr.addEventListener("click",function () {
                    if (confirm('Are you sure to rest the version to  '+info.CommitSha1+` ?`)) {
                        index.gitReset(name, info.CommitSha1);
                    }
                });
                let td = idoc.createElement("td");
                td.rowSpan = lenFile;
                td.innerText = info.CommitSha1;
                tr.appendChild(td);

                let td2 = idoc.createElement("td");
                td2.rowSpan = lenFile;
                td2.innerText = info.Author;
                tr.appendChild(td2);

                let td6 = idoc.createElement("td");
                td6.rowSpan = lenFile;
                td6.innerText = info.CommitCtx;
                tr.appendChild(td6);

                if (lenFile === 0) {
                    let td3 = idoc.createElement("td");
                    td3.innerText = "";
                    tr.appendChild(td3);

                    let td4 = idoc.createElement("td");
                    td4.innerText = "";
                    tr.appendChild(td4);
                } else {
                    let td3 = idoc.createElement("td");
                    td3.innerText = info.FileStat[0].Mode;
                    tr.appendChild(td3);

                    let td4 = idoc.createElement("td");
                    td4.innerText = info.FileStat[0].FileName;
                    tr.appendChild(td4);
                }

                let td5 = idoc.createElement("td");
                td5.rowSpan = lenFile
                td5.innerText = info.LongTime;
                tr.appendChild(td5);

                tbody.appendChild(tr);
                for (let j = 0; j < lenFile - 1; j++) {
                    let tr2 = idoc.createElement("tr");

                    let td3 = idoc.createElement("td");
                    td3.innerText = info.FileStat[j].Mode.toString();
                    tr2.appendChild(td3);

                    let td4 = idoc.createElement("td");
                    td4.innerText = info.FileStat[j].FileName;
                    tr2.appendChild(td4);
                    tbody.appendChild(tr2);
                }

            }
            gitTable.appendChild(tbody);
        })
    },

    updateSharedDirs: function(payload) {
        document.getElementById("shared").innerHTML = "";
        for (let i = 0; i < payload.length; i++) {
            index.addSharedDir(payload[i]);
            console.log(i)
        }
    },

    exploreSharedDirs: function () {
        let message = { "name": "exploreSharedDirs" };
        // Send message
        asticode.loader.show();
        astilectron.sendMessage(message, function (message) {
            asticode.loader.hide();
            // Check error
            if (message.name === "error") {
                asticode.notifier.error(message.payload);
                return
            }
            console.log(message.payload);
            document.getElementById("shared").innerHTML = "";
            for (let i = 0; i < message.payload.length; i++) {
                index.addSharedDir(message.payload[i]);
                console.log(i)
            }
        })
    },
    addSharedDir: function (name) {
        let div = document.createElement("div");
        div.className = "shared-dir";
        div.onclick = function () { index.showGitLog(name); };
        div.innerHTML = `<i class="fa fa-female"></i><span>&nbsp;` + name + `</span>`;
        document.getElementById("shared").appendChild(div)
    },
    searchFiles: function () {
        console.log("search");
        let usernameE = document.getElementById("search-username-input");
        let filenameE = document.getElementById("search-filename-input");
        let filepathE = document.getElementById("search-file-path-input");
        if (usernameE == null || filenameE == null || filepathE == null) {
            return
        }
        let username = usernameE.value;
        let filename = filenameE.value;
        let filepath = filepathE.value;
        if (username === "") {
            return
        }
        index.search(username, filename, filepath)
    },


    search: function(username, filename, filepath) {
        let message = { "name": "searchFiles", payload: { uid: username, filename: filename, filepath:filepath } };
        asticode.loader.show();
        astilectron.sendMessage(message, function (message) {
            asticode.loader.hide();
            // Check error
            if (message.name === "error") {
                asticode.notifier.error(message.payload);
                return
            }
            let idoc = document.getElementById("logInfo").contentWindow.document;

            let infoEle = idoc.getElementById("info");
            infoEle.innerHTML = "";
            let header = idoc.createElement("span");
            header.className = "search-path";
            header.id = "search-path";
            if (filepath[0] !== '/') {
                header.innerHTML = username + "/" + filepath
            } else {
                header.innerHTML = username + filepath
            }
            infoEle.appendChild(header);

            let gitTable = idoc.createElement("table");
            gitTable.className = "fileInfo table table-condensed";
            gitTable.id = "fileInfo";
            infoEle.appendChild(gitTable);

            let thead = idoc.createElement("thead");
            thead.innerHTML = `
                <tr>
                    <th style="width: 50%; text-align: left">Path</th>
                    <th style="width: 15%; text-align: left">Username</th>
                    <th style="width: 15%; text-align: left">Size</th>
                    <th style="width: 20%; text-align: left">MTime</th>
                </tr>
            `;

            let tbody = idoc.createElement("tbody");

            gitTable.appendChild(thead);
            gitTable.appendChild(tbody);

            let ps = filepath.split("/");
            console.log("filepath:", filepath, ps.length);
            if (filepath !== "" && ps.length >= 1) {
                ps = ps.slice(0, ps.length-1);
                let tr = idoc.createElement("tr");
                tr.innerHTML = "<td><i class=\"fa fa-folder\"></i><span>&nbsp;..</span></td>"+
                    "<td></td>"+
                    "<td></td>";
                tr.addEventListener("click", function f() {
                    index.search(username, filename, ps.join("/"))
                });
                tbody.appendChild(tr);
            }


            for (let i = 0; i < message.payload.length; i++) {
                let info = message.payload[i];
                console.log(info);
                let tr = idoc.createElement("tr");
                if (info.type === true) {
                    tr.innerHTML = "<td><i class=\"fa fa-folder\"></i><span>&nbsp;" + info.path + "</span></td>"+
                        "<td>"+info.uid+"</td>"+
                        "<td>"+info.size+"</td>"+
                        "<td>"+info.m_time+"</td>";
                    tr.addEventListener("click", function f() {
                        index.search(username, filename, info.path)
                    });
                } else  {
                    tr.innerHTML = "<td><i class=\"fa fa-file\"></i><span>&nbsp;" + info.path + "</span></td>"+
                        "<td>"+info.uid+"</td>"+
                        "<td>"+info.size+"</td>"+
                        "<td>"+info.m_time+"</td>";
                }
                tbody.appendChild(tr)
            }
        });
    },

    mount: function (uid, point) {
        if (typeof uid === "undefined" || uid === "") {
            return
        }
        let message = { "name": "mount", payload: {uid:uid, point:point} };
        console.log(message);
        // Send message
        asticode.loader.show();
        astilectron.sendMessage(message, function (message) {
            asticode.loader.hide();
            if (typeof message === "undefined") {
                return
            }
            // Check error
            if (message.name === "error") {
                asticode.notifier.error(message.payload);
                return
            }
            // mount show
            let mounted = document.getElementById("mounted");
            let dev = document.createElement("dev");
            dev.className = "mountedOne";
            dev.id = "mounted-"+uid;
            mounted.appendChild(dev);

            let span = document.createElement("span");
            span.className = "mounted-username";
            span.textContent = uid + " \\ " + point;

            let button = document.createElement("button");
            button.className = "unmount-button";
            button.innerText = "Unmount";
            button.addEventListener("click", function () {
                index.unmount(uid)
            });

            dev.appendChild(span);
            dev.appendChild(button)

        })
    },
    unmount: function(uid){
        if (typeof uid === "undefined" || uid === "") {
            return
        }
        let message = { "name": "unmount", payload: uid };
        console.log(message);
        // Send message
        asticode.loader.show();
        astilectron.sendMessage(message, function (message) {
            asticode.loader.hide();
            console.log("unmount", message);
            // Check error
            if (typeof message !== "undefined" && message.name === "error") {
                asticode.notifier.error(message.payload);
                return
            }
            // remove dev
            let mounted = document.getElementById("mounted");
            let d = document.getElementById("mounted-"+uid);
            console.log("remove", mounted, d);
            mounted.removeChild(d);
        })
    },
    confirmMount: function() {
        let usernameE = document.getElementById("search-username-input");
        if (usernameE == null) {
            return
        }
        if (usernameE.value === "") {
            return
        }
        let pointE = document.getElementById("mount-point-input");
        if (pointE == null) {
            return
        }
        if (pointE.value === "") {
            return
        }
        let username = usernameE.value;
        if (confirm('Are you sure to mount '+name+` at+`+pointE.value+`+?`)) {
            index.mount(username, pointE.value)
        }
    },

    login: function () {
        let ud = document.getElementById("input-username")
        if (ud == null) {
            return
        }
        let username = ud.value;
        let password = document.getElementById("input-password").value;
        if (username === "" || password === "") {
            return
        }
        let message = { "name": "login" };
        message.payload = { "username": username, "password": password };
        asticode.loader.show();
        astilectron.sendMessage(message, function (message) {
            asticode.loader.hide();
            console.log(message);
            // Check error
            if (typeof message !== "undefined" && message.name === "error") {
                asticode.notifier.error(message.payload);
                return
            }
            asticode.loader.show();
            window.location.replace("index.html");
            document.title = username;
            asticode.loader.hide();
        })

    },
    register: function () {
        let ud = document.getElementById("input-username")
        if (ud == null) {
            return
        }
        let username = ud.value;
        let password = document.getElementById("input-password").value;
        let message = { "name": "register" };
        if (username === "" || password === "") {
            return
        }

        message.payload = { "username": username, "password": password };
        asticode.loader.show();
        astilectron.sendMessage(message, function (message) {
            asticode.loader.hide();

            console.log(message);
            // Check error
            if (typeof message !== "undefined" && message.name === "error") {
                asticode.notifier.error(message.payload);
                return
            }

            asticode.loader.show();
            window.location.replace("login.html");
            asticode.loader.hide();
        })
    },

    gitReset: function(name, sha1) {
        if (typeof sha1 === "undefined" || typeof name === "undefined") {
            return
        }
        if (sha1 === "" || name === "") {
            return
        }
        let msg = {name:"gitRest", payload:{name:name, sha1:sha1}};
        asticode.loader.show();
        astilectron.sendMessage(msg, function (message) {
            // Check error
            if (typeof message == "undefined") {
                index.showGitLog(name)
                return
            }
            if (message.name === "error") {
                asticode.notifier.error(message.payload);
                return
            }
            index.showGitLog(name)
        })
    },
    listen: function () {
        astilectron.onMessage(function (message) {
            switch (message.name) {
                case "about":
                    index.about(message.payload);
                    return { payload: "payload" };
                case "config":
                    console.log(message.payload);
                    index.configShow();
                    break;
                case "replace":
                    console.log(message.payload);
                    window.location.replace(message.payload);
                    break;
                case "updateSharedDirs":
                    index.updateSharedDirs(message.payload);
                    break;
            }
        });
    },
    configShow:function () {
        window.open("config.html")
    },
};