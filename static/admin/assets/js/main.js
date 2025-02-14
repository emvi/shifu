(function() {
    document.addEventListener("htmx:afterProcessNode", details => {
        if (details.target.classList.contains("shifu-window")) {
            const title = details.target.querySelector(".shifu-window-title");

            if (title) {
                const settings = localStorage.getItem(details.target.id);

                if (settings) {
                    const position = JSON.parse(settings);
                    details.target.style.top = position.top;
                    details.target.style.left = position.left;
                }

                title.addEventListener("mousedown", startDrag);
                title.addEventListener("mouseup", endDrag);
            }
        }
    });
    document.addEventListener("htmx:beforeCleanupElement", cleanup);
    document.addEventListener("htmx:trigger", details => {
        if (details.target.classList.contains("shifu-window-title-close")) {
            cleanup(details);
            details.target.parentNode.parentNode.remove();
        }
    });
    window.addEventListener("mousemove", drag);
    const save = debounce(target => {
        localStorage.setItem(target.id, JSON.stringify({
            top: target.style.top,
            left: target.style.left
        }));
    });
    let dragging = null;
    let mouseX, mouseY, zIndex = 0;

    function startDrag(e) {
        dragging = e.target.parentNode;

        if (dragging.style.zIndex > zIndex) {
            zIndex = dragging.style.zIndex;
        }

        zIndex++;
        dragging.style.zIndex = zIndex;
        mouseX = e.clientX;
        mouseY = e.clientY;
    }

    function endDrag() {
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
            save(dragging);
        }
    }

    function cleanup(details) {
        if (details.target.classList.contains("shifu-window")) {
            const title = details.target.querySelector(".shifu-window-title");

            if (title) {
                title.removeEventListener("mousedown", startDrag);
                title.removeEventListener("mouseup", endDrag);
            }
        }
    }

    function debounce(func, timeout = 300) {
        let timer;
        return (...args) => {
            clearTimeout(timer);
            timer = setTimeout(() => {
                func.apply(this, args);
            }, timeout);
        };
    }
})();
