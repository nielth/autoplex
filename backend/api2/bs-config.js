module.exports = {
    proxy: "localhost:8081",  // Proxy your local server
    files: ["./app/templates/**/*.html"],  // Watch for changes in the templates folder
    notify: false,  // Disable the "BrowserSync Enabled" popup
    open: false  // Prevents BrowserSync from opening a new tab automatically
};
