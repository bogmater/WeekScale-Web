(() => {
    const switcher = document.querySelector(".platform-switcher");
    if (!switcher) {
        return;
    }

    const buttons = Array.from(switcher.querySelectorAll("button[data-platform]"));
    const images = Array.from(document.querySelectorAll("img[data-ios-src]"));
    const storageKey = "weekscale.platform";

    const setPlatform = (platform) => {
        const active = platform === "ios" ? "ios" : "android";
        buttons.forEach((button) => {
            const isActive = button.dataset.platform === active;
            button.classList.toggle("is-active", isActive);
            button.setAttribute("aria-pressed", String(isActive));
        });
        images.forEach((img) => {
            const useIos = active === "ios";
            img.src = useIos ? img.dataset.iosSrc : img.dataset.androidSrc;
            img.srcset = useIos ? img.dataset.iosSrcset : img.dataset.androidSrcset;
        });
        try {
            window.localStorage.setItem(storageKey, active);
        } catch (error) {
            // Storage may be unavailable; the toggle still works for this page.
        }
    };

    buttons.forEach((button) => {
        button.addEventListener("click", () => setPlatform(button.dataset.platform));
    });

    let initial = "android";
    try {
        const saved = window.localStorage.getItem(storageKey);
        if (saved === "ios" || saved === "android") {
            initial = saved;
        }
    } catch (error) {
        // Fall back to Android.
    }
    setPlatform(initial);
})();
