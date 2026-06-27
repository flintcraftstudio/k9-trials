// avatar.js — client behaviour for the AvatarUpload control (components.AvatarUpload).
//
// Pure DOM, no framework: the control is used inside htmx-swapped forms
// (the profile editor and dog form re-render on validation error), and
// Alpine is not re-initialised in swapped content in this app. Inline
// onchange/onclick attributes plus these global helpers survive swaps
// because the functions live on window and the attributes are re-bound by
// the browser whenever the markup is parsed.
//
// Storage/upload is intentionally not wired anywhere — this only manages
// the in-page affordance: choose a file (with live preview) or clear the
// current photo. The chosen file posts under the field's name; a companion
// hidden "<name>_remove" field flips to "true" when the photo is cleared.
(function () {
  if (window.k9Avatar) return;

  function root(el) {
    return el.closest("[data-avatar]");
  }

  window.k9Avatar = {
    // open the native file picker for this control
    open: function (btn) {
      var file = root(btn).querySelector("[data-avatar-file]");
      if (file) file.click();
    },

    // a file was chosen: show a live preview and reset the remove flag
    pick: function (input) {
      var r = root(input);
      var file = input.files && input.files[0];
      if (!file) return;
      var img = r.querySelector("[data-avatar-img]");
      var ph = r.querySelector("[data-avatar-placeholder]");
      var rm = r.querySelector("[data-avatar-remove]");
      var flag = r.querySelector("[data-avatar-flag]");
      if (img) {
        img.src = URL.createObjectURL(file);
        img.hidden = false;
      }
      if (ph) ph.hidden = true;
      if (rm) rm.hidden = false;
      if (flag) flag.value = "false";
    },

    // clear the photo: drop any chosen file, restore the initials
    // placeholder, and mark the photo for removal
    remove: function (btn) {
      var r = root(btn);
      var img = r.querySelector("[data-avatar-img]");
      var ph = r.querySelector("[data-avatar-placeholder]");
      var file = r.querySelector("[data-avatar-file]");
      var flag = r.querySelector("[data-avatar-flag]");
      if (file) file.value = "";
      if (img) {
        img.src = "";
        img.hidden = true;
      }
      if (ph) ph.hidden = false;
      btn.hidden = true;
      if (flag) flag.value = "true";
    },
  };
})();
