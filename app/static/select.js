$(document).ready(function () {
    const currentPage = window.location.pathname;

    let active = true
    $('.header .header-right a').each(function(){
        if($(this).attr('href') == currentPage){
            $(this).addClass('active');
            active = false
        } 
    });
    if (active){
        
        home = document.getElementsByClassName("logo home");
        console.log(home)
        home[0].className += " active"
    }
});