(function () {
    document.body.addEventListener("tt:toast", function (event) {
        const detail = event.detail || {};
        const message = typeof detail.message === "string" ? detail.message : "";
        showToast(detail.type, message);
    });

    document.body.addEventListener("htmx:responseError", function (event) {
        const requestConfig = event.detail && event.detail.requestConfig;
        if (!requestConfig) {
            return;
        }
        const xhr = event.detail && event.detail.xhr;
        const responseMessage = xhr && typeof xhr.responseText === "string"
            ? xhr.responseText.trim()
            : "";
        showErrorToast(responseMessage);
    });

    document.body.addEventListener("htmx:sendError", function (event) {
        const requestConfig = event.detail && event.detail.requestConfig;
        if (!requestConfig) {
            return;
        }
        showErrorToast("Request could not be sent.");
    });
})();

function showSuccessToast(message) {
    showToast("success", message || "Success.");
}

function showErrorToast(message) {
    showToast("error", message || "An error occurred.");
}


function showToast(type, message) {
    // Early exit without message
    if (!message) {
        return;
    }
    // Define the template
    var templateName = type === "success" ? "tt-success-toast-template" : "tt-error-toast-template";
    var containerEl = document.getElementById("tt-toast-container");
    var templateEl = document.getElementById(templateName);
    if (!containerEl || !templateEl) {
        return;
    }
    // Clone the template
    const toastEl = templateEl.cloneNode(true);
    toastEl.removeAttribute("id");
    toastEl.classList.remove("d-none");
    // Set the message
    const toastBodyEl = toastEl.querySelector(".toast-body");
    if (!toastBodyEl) {
        return;
    }
    toastBodyEl.textContent = message;
    // Append the toast
    containerEl.appendChild(toastEl);
    // Make sure to remove the toast element from the DOM when it's hidden
    toastEl.addEventListener("hidden.bs.toast", function () {
        toastEl.remove();
    });
    // Show the toast
    const toast = new bootstrap.Toast(toastEl);
    toast.show();
}
