(function() {
    const storageKey = "shifu-media-view";

    function setView(view) {
        if (view === "tiles") {
            document.documentElement.setAttribute("data-shifu-media-tiles", "");
        } else {
            document.documentElement.removeAttribute("data-shifu-media-tiles");
        }
    }

    setView(localStorage.getItem(storageKey));

    document.addEventListener("click", e => {
        const button = e.target.closest ? e.target.closest("[data-shifu-media-view]") : null;

        if (button) {
            e.preventDefault();
            e.stopPropagation();
            const view = button.getAttribute("data-shifu-media-view");
            localStorage.setItem(storageKey, view);
            setView(view);
            return;
        }

        // in the tile view, clicking a tile toggles its checkbox
        const tile = document.documentElement.hasAttribute("data-shifu-media-tiles") && e.target.closest ?
            e.target.closest(".shifu-media-table tbody tr") : null;

        if (tile && !e.target.closest("input, button, a, label")) {
            const checkbox = tile.querySelector(".shifu-multi-file-select");

            if (checkbox) {
                checkbox.checked = !checkbox.checked;
            }
        }
    });
})();
