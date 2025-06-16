window.shifuAddElement = function(parent, position) {
    const tpl = document.getElementById("shifu-new-element");

    if (tpl) {
        const slot = document.querySelector(`slot[name='${parent}/${position}']`);

        if (slot) {
            const parent = slot.parentElement;
            parent.appendChild(tpl.content);
            slot.remove();
            parent.appendChild(slot);
        } else {
            document.body.appendChild(tpl.content);
        }
    } else {
        console.error("New element template not found");
        location.reload();
    }
}

window.shifuMoveElement = function(selector, direction) {
    const move = document.querySelector(`[data-shifu-element='${selector}']`);

    if (move) {
        if (direction === "up") {
            let previous = move.previousElementSibling;

            while (!previous.hasAttribute("data-shifu-element")) {
                previous = previous.previousElementSibling;
            }

            if (!previous) {
                console.error("Previous element not found", selector);
                location.reload();
                return;
            }

            const id = previous.getAttribute("data-shifu-element");
            previous.setAttribute("data-shifu-element", move.getAttribute("data-shifu-element"));
            move.setAttribute("data-shifu-element", id);
            move.parentNode.insertBefore(move, previous);
        } else {
            let next = move.nextElementSibling;

            while (!next.hasAttribute("data-shifu-element")) {
                next = next.nextElementSibling;
            }

            if (!next) {
                console.error("Next element not found", selector);
                location.reload();
                return;
            }

            const id = next.getAttribute("data-shifu-element");
            next.setAttribute("data-shifu-element", move.getAttribute("data-shifu-element"));
            move.setAttribute("data-shifu-element", id);
            move.parentNode.insertBefore(next, move);
        }
    } else {
        console.error("Element to move not found", selector, direction);
        location.reload();
    }
}

window.shifuDeleteElement = function(selector) {
    const element = document.querySelector(`[data-shifu-element='${selector}']`);

    if (element) {
        element.remove();
    } else {
        console.error("Element to delete not found", selector);
        location.reload();
    }
}
