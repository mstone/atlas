/*
 * ext-server_atlas.js
 *
 * Licensed under the MIT License
 *
 * Copyright(c) 2010 Alexis Deveria
 *              2011 MoinMoin:ReimarBauer
 *                   adopted for moinmoins item storage. it sends in one post png and svg data
 *                   (I agree to dual license my work to additional GPLv2 or later)
 *              2013 Akamai Technologies, Inc.
 *
 */

svgEditor.addExtension("server_opensave", {
  callback: function() {
    var uploadIframe = $('<iframe name="output_frame" src="#"/>').hide().appendTo('body');
    svgEditor.setCustomHandlers({
      save: function(win, data) {
        //var formTarget = window.parent.document.location;
        var formTarget = window.document.location;
        var svg = "<?xml version=\"1.0\"?>\n" + data;
        var b64_svg = svgedit.utilities.encode64(svg);
        var form = $('<form>').attr({
          method: 'post',
          action: formTarget,
          target: 'output_frame'
        })  .append('<input type="hidden" name="filepath" value="' + b64_svg + '">')
            .append('<input type="hidden" name="filename" value="drawing.svg">')
            .append('<input type="hidden" name="contenttype" value="application/x-svgdraw">')
            .appendTo('body')
            .submit().remove();
        alert("Saved!");
      },
    });
  },
});

