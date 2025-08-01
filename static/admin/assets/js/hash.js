export function simpleHash(str) {
    let hash = 729790;

    for (let i = 0; i < str.length; i++) {
        hash = ((hash << 5) + hash) + str.charCodeAt(i);
    }

    return (hash >>> 0).toString(16).slice(0, 8); // Convert to positive hex
}
