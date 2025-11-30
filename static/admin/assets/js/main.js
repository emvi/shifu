import {debounce} from "./debounce";

(function() {
    const save = debounce(target => {
        localStorage.setItem(target.id, JSON.stringify({
            top: target.style.top,
            left: target.style.left
        }));
    });
    let dragging = null;
    let mouseX, mouseY, zIndex = 100_000;
    document.addEventListener("htmx:beforeRequest", e => {
        const id = e.target.getAttribute("data-window");

        if (id) {
            const element = document.getElementById(id);

            if (element) {
                cleanup(e);
                element.remove();
                e.preventDefault();
            }
        }
    });
    document.addEventListener("htmx:afterProcessNode", e => {
        if (e.target.classList.contains("shifu-window") || e.target.classList.contains("shifu-window-overlay")) {
            let target = e.target;

            if (target.classList.contains("shifu-window-overlay")) {
                target = target.children[0];
            }

            document.querySelectorAll(".shifu-window").forEach(e => e.classList.remove("shifu-window-active"));
            target.classList.add("shifu-window-active");
            zIndex++;
            target.style.zIndex = zIndex;
            const title = target.querySelector(".shifu-window-title");

            if (title) {
                const settings = localStorage.getItem(target.id);

                if (settings) {
                    const position = JSON.parse(settings);
                    target.style.top = position.top;
                    target.style.left = position.left;
                }

                title.addEventListener("mousedown", startDrag);
                title.addEventListener("mouseup", endDrag);
            }
        }
    });
    document.addEventListener("htmx:beforeCleanupElement", cleanup);
    document.addEventListener("htmx:trigger", e => {
        if (e.target.classList.contains("shifu-window-close")) {
            cleanup(e);
            let window = e.target;

            while (!window.classList.contains("shifu-window")) {
                window = window.parentNode;
            }

            if (window.parentNode && window.parentNode.classList.contains("shifu-window-overlay")) {
                window.parentNode.remove();
            }

            window.remove();
        }
    });
    window.addEventListener("mousemove", drag);
    window.addEventListener("mouseup", endDrag);

    function startDrag(e) {
        document.querySelectorAll(".shifu-window").forEach(e => e.classList.remove("shifu-window-active"));
        dragging = e.target.parentNode;

        if (dragging.style.zIndex > zIndex) {
            zIndex = dragging.style.zIndex;
        }

        zIndex++;
        dragging.style.zIndex = zIndex;
        dragging.classList.add("shifu-window-active");
        mouseX = e.clientX;
        mouseY = e.clientY;
    }

    function endDrag() {
        dragging = null;
    }

    function drag(e) {
        if (e.clientX < 0 || e.clientX > window.innerWidth || e.clientY < 0 || e.clientY > window.innerHeight) {
            endDrag();
            return;
        }

        if (dragging) {
            const deltaX = mouseX - e.clientX;
            const deltaY = mouseY - e.clientY;
            let x = dragging.offsetLeft - deltaX;
            let y = dragging.offsetTop - deltaY;
            mouseX = e.clientX;
            mouseY = e.clientY;
            const rect = dragging.getBoundingClientRect();

            if (deltaX > 0 && rect.left <= 0) {
                x = 0;
            } else if (deltaX < 0 && rect.right >= window.innerWidth) {
                x = window.innerWidth - rect.width;
            }

            if (deltaY > 0 && rect.top <= 0) {
                y = 0;
            } else if (deltaY < 0 && rect.bottom >= window.innerHeight) {
                y = window.innerHeight - rect.height;
            }

            dragging.style.top = y + "px";
            dragging.style.left = x + "px";
            save(dragging);
        }
    }

    function cleanup(details) {
        if (details.target.classList && details.target.classList.contains("shifu-window")) {
            const title = details.target.querySelector(".shifu-window-title");

            if (title) {
                title.removeEventListener("mousedown", startDrag);
                title.removeEventListener("mouseup", endDrag);
            }
        }
    }

    document.addEventListener("htmx:afterRequest", e => {
        const target = e.target.getAttribute("data-window");

        if (target) {
            let window = document.querySelector(target);

            if (!window) {
                return;
            }

            while (!window.classList.contains("shifu-window")) {
                window = window.parentNode;
            }

            const title = window.querySelector(".shifu-window-title");

            if (title) {
                title.removeEventListener("mousedown", startDrag);
                title.removeEventListener("mouseup", endDrag);
            }

            if (window.parentNode && window.parentNode.classList.contains("shifu-window-overlay")) {
                window.parentNode.remove();
            }

            window.remove();
        }
    });
})();

import "./trix";
import "./tree";
import "./page";
