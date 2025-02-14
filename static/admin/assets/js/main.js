(function() {
    document.addEventListener("htmx:afterSwap", details => {
        if (details.target.classList.contains("shifu-window")) {
            const title = details.target.querySelector(".shifu-window-title");

            if (title) {
                title.addEventListener("mousedown", startDrag);
                title.addEventListener("mouseup", endDrag);
            }
        }
    });
    document.addEventListener("htmx:beforeCleanupElement", details => {
        if (details.target.classList.contains("shifu-window")) {
            const title = details.target.querySelector(".shifu-window-title");

            if (title) {
                title.removeEventListener("mousedown", startDrag);
                title.removeEventListener("mouseup", endDrag);
            }
        }
    });
    window.addEventListener("mousemove", drag);
    let dragging = null;
    let mouseX, mouseY;

    function startDrag(e) {
        dragging = e.target.parentNode;
        mouseX = e.clientX;
        mouseY = e.clientY;
    }

    function endDrag(e) {
        dragging = null;
    }

    function drag(e) {
        if (dragging) {
            const x = mouseX - e.clientX;
            const y = mouseY - e.clientY;
            mouseX = e.clientX;
            mouseY = e.clientY;
            dragging.style.top = (dragging.offsetTop - y) + "px";
            dragging.style.left = (dragging.offsetLeft - x) + "px";
        }
    }
})();
