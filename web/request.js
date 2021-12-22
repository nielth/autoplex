/*
#############################
#                           #
#   WARNING! I suck at JS   #
#                           #
#############################
*/

const ip_addr = "10.13.13.3"

const settings = {
    "async": false,
    "crossDomain": true,
    "url": "",
    "method": "GET",
};

$("#title").keyup(function (event) {
    if (event.keyCode === 13) {
        $("#titleButton").click();
    }
});


document.getElementById("titleButton").addEventListener("click", function () {
    const container = document.getElementById('imageContainer');
    // Remove previous search
    container.textContent = '';
    // Selecting the input element and get its value
    const inputVal = document.getElementById("title").value;

    const loader = document.getElementById('loader-div');

    loader.style.display = "inline";

    // Removes unnecessary spaces
    var newStr = inputVal.replaceAll(" ", "%");
    while (newStr[newStr.length - 1] === "%") {
        newStr = newStr.slice(0, -1);
    }

    settings.url = `http://${ip_addr}:5000/imdb/movie/${newStr}/`

    $.getJSON(settings, function (result) {
        for (let i = 0; i < Object.keys(result).length; i++) {
            if (result[i]['cover-url'].includes("https://m.media-amazon.com")) {
                const element = document.createElement("div");
                element.id = "title-choose"

                // Find the image
                const img = document.createElement('img');
                img.src = result[i]['cover-url'];
                img.alt = result[i]['movie-id'];
                img.onclick = function () {
                    getMagnet(event)
                };

                // Find the title
                const para = document.createElement("p");
                para.innerText = result[i]['title'];
                container.appendChild(para);

                // Append image and title into div
                container.appendChild(element).appendChild(img);
                container.appendChild(element).appendChild(para);
            }
        }
    });
    loader.style.display = "none";
})

$(document).ready(function () {
    $('#title-choose').on('click', getMagnet);
});


function getMagnet(event) {
    const container = document.getElementById('imageContainer');
    container.textContent = '';

    const alt = event.target.alt;
    const img = event.target.src;
    const image = document.createElement("img");
    image.src = img;
    image.style.width = "13%";
    image.style.padding = "67px 0 36px 0";
    container.appendChild(image);

    settings.url = `http://${ip_addr}:5000/torrent/${alt}/`

    const loader = document.getElementById('loader-div');
    loader.style.display = "inline";

    $.getJSON(settings, function (result) {
        const element = document.createElement("div");
        const table = document.createElement("table");
        const tr = document.createElement("tr");
        addHeaderTable(container, "Name of file");
        addHeaderTable(container, "Seeders");
        addHeaderTable(container, "Size");
        addHeaderTable(container, "Download");

        function addHeaderTable(container, content) {
            const th = document.createElement("th");
            th.innerText = content;
            container.appendChild(th);
            container.appendChild(element).appendChild(table).appendChild(tr).appendChild(th);
        }

        function addTable(container, tr, content) {
            const td = document.createElement("td");
            td.innerText = content;
            container.appendChild(td);
            container.appendChild(element).appendChild(table).appendChild(tr).appendChild(td);
        }

        for (let i = 0; i < Object.keys(result).length; i++) {
            const tr = document.createElement("tr");
            const td = document.createElement("td");
            addTable(container, tr, result[i]['filename']);
            addTable(container, tr, result[i]['seeders']);
            addTable(container, tr, result[i]['size'] + " GB");
            const a = document.createElement("a");
            const img = document.createElement('img');
            img.src = "https://dyncdn.me/static/20/img/magnet.gif";
            img.alt = result[i]['magnet']
            img.onclick = function () {
                torrentDownload(this.alt);
            };
            container.appendChild(element).appendChild(table).appendChild(tr).appendChild(td).appendChild(a).appendChild(img)
        }
    });
    loader.style.display = "none";

}

function torrentDownload(magnet) {
    console.log(magnet)
    const magnet_link = {'magnet': magnet}
    $.ajax({
        url: `http://${ip_addr}:5000/movie/`,
        "async": false,
        type: 'POST',
        dataType: 'json',
        data: magnet_link,
        success: function () {
            alert("Success!")
        },
    });
}