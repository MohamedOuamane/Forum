function togglePasswordLog() {
    const field = document.getElementById("password-log");
    if (field) {
        field.type = field.type === "password" ? "text" : "password";
    }
}

function togglePasswordReg() {
    const field = document.getElementById("password-reg");
    if (field) {
        field.type = field.type === "password" ? "text" : "password";
    }
}

// Generic toggle used by change password popup
function togglePassword(inputId) {
    const field = document.getElementById(inputId);
    if (field) {
        field.type = field.type === "password" ? "text" : "password";
    }
}

function closeChangePasswordPopup() {
    const el = document.getElementById('changePasswordPopup');
    if (el) el.style.display = 'none';
}

// Only attach listener if button exists (profile page only)
const changePasswordBtn = document.getElementById('change-password');
if (changePasswordBtn) {
    changePasswordBtn.addEventListener('click', function() {
        const popup = document.getElementById('changePasswordPopup');
        if (popup) popup.style.display = 'block';
    });
}