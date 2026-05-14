// Like and dislike counts are handled server-side.
// After every like/dislike POST the page redirects back with real counts from the DB —
// no client-side count manipulation needed or wanted here.
//
// When you're ready to make likes feel instant without a page reload,
// switch the like/dislike buttons to use fetch() and update the count in the
// response — but that also requires changes to your Go handlers to return JSON.

function focusTextarea() {
    const el = document.getElementById("CommentArea");
    if (el) {
        el.focus();
        el.scrollIntoView({ behavior: "smooth", block: "center" });
    }
}