// Need fix
$("#title").keyup(function (event) {
    if (event.keyCode === 13) {
        $("#titleButton").click();
    }
});

const settings = {
    "async": true,
    "crossDomain": true,
    "url": "",
    "method": "GET",
};

function getTitleValue(name, value) {
    const container = document.getElementById('imageContainer');
    // Remove previous search
    container.textContent = '';
    // Selecting the input element and get its value
    const inputVal = document.getElementById("title").value;

    // Removes unnecessary spaces
    var newStr = inputVal.replaceAll(" ", "%");
    while (newStr[newStr.length - 1] === "%") {
        newStr = newStr.slice(0, -1);
    }

    settings.url = `http://127.0.0.1:5000/imdb/${newStr}/`

    $.getJSON(settings, function (result) {
        for (let i = 0; i < Object.keys(result).length; i++) {
            const element = document.createElement("div");

            // Find the image
            const img = document.createElement('img');
            img.src = result[i]['cover-url'];
            img.alt = result[i]['movie-id'];
            img.onclick = function () {
                titleValue(event)
            };

            // Find the title
            const para = document.createElement("p");
            para.innerText = result[i]['title'];
            container.appendChild(para);

            // Append image and title into div
            container.appendChild(element).appendChild(img);
            container.appendChild(element).appendChild(para);
        }
    });
}

function titleValue(event) {
    const container = document.getElementById('imageContainer');
    container.textContent = '';

    const alt = event.target.alt;

    settings.url = `http://127.0.0.1:5000/torrent/${alt}/`
    const element = document.createElement("div");
    const table = document.createElement("table");
    const tr = document.createElement("tr")
    addHeaderTable(container, "Name of file")
    addHeaderTable(container, "Seeders")
    addHeaderTable(container, "Size")
    addHeaderTable(container, "Download")

    function addHeaderTable(container, content){
        const th = document.createElement("th");
        th.innerText = content;
        container.appendChild(th)
        container.appendChild(element).appendChild(table).appendChild(tr).appendChild(th)
    }


    $.getJSON(settings, function (result) {
        for (let i = 0; i < Object.keys(result).length; i++) {
            const tr = document.createElement("tr")
            addTable(container, tr, result[i]['filename'])
            addTable(container, tr, result[i]['seeders'])
            addTable(container, tr, result[i]['size'] + " GB")
            const img = document.createElement('img');
            img.src = "https://dyncdn.me/static/20/img/magnet.gif";
            img.style.width = "12px";
            img.onclick = function () {
                titleValue(event)
            };
            container.appendChild(element).appendChild(table).appendChild(tr).appendChild(img)
        }
    });

    function addTable(container, tr, content){
    const td = document.createElement("td")
    td.innerText = content;
    container.appendChild(td);
    container.appendChild(element).appendChild(table).appendChild(tr).appendChild(td);
}
}