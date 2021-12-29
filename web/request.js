/*
#############################
#                           #
#   WARNING! I suck at JS   #
#                           #
#############################
*/

const ip_addr = "localhost";
let movie = false;
let series = false;

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

$('input[type="checkbox"]').on('change', function () {
    $('input[type="checkbox"]').not(this).prop('checked', false);
});


document.getElementById("titleButton").addEventListener("click", function () {
    const container = document.getElementById('imageContainer');
    if (document.getElementById("movie").checked) {
        movie = true;
        series = false;
    } else if (document.getElementById("series").checked) {
        movie = false;
        series = true;
    }
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

    if (movie) {
        settings.url = `http://${ip_addr}:5000/imdb/movie/${newStr}/`;
    } else if (series) {
        settings.url = `http://${ip_addr}:5000/imdb/series/${newStr}/`;
    } else {
        alert("Choose movie or tv series");
    }

    $.getJSON(settings, function (result) {
        const lenResults = Object.keys(result).length
        for (let i = 0; i < lenResults; i++) {
            if (result[i]['cover-url'].includes("https://m.media-amazon.com")) {
                const element = document.createElement("div");
                element.id = "title-choose";

                // Find the image
                const img = document.createElement('img');
                img.src = result[i]['cover-url'];
                img.alt = result[i]['movie-id'];
                element.onclick = function () {
                    getMagnet(event);
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
    image.style.padding = "25px 0 36px 0";
    container.appendChild(image);

    if (movie) {
        settings.url = `http://${ip_addr}:5000/torrent/movie/${alt}/`;
    } else if (series) {
        settings.url = `http://${ip_addr}:5000/torrent/series/${alt}/`;
    }

    const loader = document.getElementById('loader-div');
    loader.style.display = "inline";

    $.getJSON(settings, function (result) {
        const element = document.createElement("div");
        const table = document.createElement("table");
        table.id = "myTable";
        const tr = document.createElement("tr");
        addHeaderTable(container, "Name of file", 0);
        addHeaderTable(container, "Seeders", 1);
        addHeaderTable(container, "Size", 2);
        addHeaderTable(container, "Download", 3);

        function addHeaderTable(container, content, position) {
            const th = document.createElement("th");
            if (position === 0 || position === 1 || position === 2) {
                th.onclick = function () {
                    sortTable(position)
                }
            }
            th.innerText = content;
            container.appendChild(th);
            container.appendChild(element).appendChild(table).appendChild(tr).appendChild(th);
        }

        const lenResults = Object.keys(result).length;
        for (let i = 0; i < lenResults; i++) {
            const tr = document.createElement("tr");
            const td = document.createElement("td");
            addTable(tr, result[i]['filename']);
            addTable(tr, result[i]['seeders']);
            addTable(tr, result[i]['size'] + " GB");
            const a = document.createElement("a");
            const img = document.createElement('img');
            img.src = "https://dyncdn.me/static/20/img/magnet.gif";
            img.alt = result[i]['magnet']
            img.onclick = function () {
                torrentDownload(this.alt);
            };
            container.appendChild(element).appendChild(table).appendChild(tr).appendChild(td).appendChild(a).appendChild(img);
        }

        function addTable(tr, content) {
            const td = document.createElement("td");
            td.innerText = content;
            container.appendChild(td);
            container.appendChild(element).appendChild(table).appendChild(tr).appendChild(td);
        }
    });
    loader.style.display = "none";
}


// Add method to authenticate magnet being POST-ed. Save title, request all torrents and compare each magnet to POST magnet
function torrentDownload(magnet) {
    const magnet_link = {
        'magnet': magnet,
        'title': save_title
    };
    let urlCategory = "";
    if (movie) {
        urlCategory = `http://${ip_addr}:5000/movie/`;
    } else if (series) {
        urlCategory = `http://${ip_addr}:5000/series/`;
    }
    $.ajax({
        url: urlCategory,
        "async": false,
        type: 'POST',
        dataType: 'json',
        data: magnet_link,
        success: function () {
            alert("Success!");
        },
    });
}