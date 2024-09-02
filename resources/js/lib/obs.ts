let options = {
    // root: document.querySelector('#app'),
    rootMargin: "0px",
    threshold: 1.0,
};
let observer = new IntersectionObserver((entries) => {
    entries.forEach(i => {
        if (i.intersectionRatio > 0) {
            // 元素进入可视区域
            const event = new CustomEvent('visible', {
                detail: {
                    targetElement: i.target
                }
            });
            i.target.dispatchEvent(event);
        }
    });
}, options);

export default observer;