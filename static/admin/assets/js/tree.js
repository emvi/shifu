window.shifuTree = function(id) {
    const tree = document.getElementById(id);

    if (!tree) {
        console.error("Shifu tree element not found", id)
        return;
    }

    let state = new Map;
    const stateFromLocalStorage = localStorage.getItem(id);

    if (stateFromLocalStorage) {
        state = new Map(JSON.parse(stateFromLocalStorage));
    }

    tree.querySelectorAll("details").forEach(element => {
        const entry = element.getAttribute("data-entry");

        if (state.get(entry)) {
            element.setAttribute("open", "");
        }

        element.addEventListener("click", event => {
            let target = event.target;

            while (!target.hasAttribute("data-entry")) {
                target = target.parentNode;
            }

            const entry = target.getAttribute("data-entry");
            const open = target.hasAttribute("open");

            if (!open) {
                state.set(entry, true);
            } else {
                state.delete(entry);
            }

            localStorage.setItem(id, JSON.stringify(Array.from(state.entries())));
        });
    });
}
