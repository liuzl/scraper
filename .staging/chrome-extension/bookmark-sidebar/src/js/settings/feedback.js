($ => {
    "use strict";

    window.FeedbackHelper = function (s) {

        /**
         *
         * @returns {Promise}
         */
        this.init = async () => {
            initEvents();

            if (s.serviceAvailable === false) {
                s.opts.elm.feedback.form.addClass(s.opts.classes.hidden);

                $("<p />")
                    .addClass(s.opts.classes.error)
                    .html(s.helper.i18n.get("status_feedback_unavailable_desc") + "<br />")
                    .append("<a href='mailto:feedback@blockbyte.de'>feedback@blockbyte.de</a>")
                    .insertAfter(s.opts.elm.feedback.form);
            }
        };

        /**
         * Initialises the eventhandlers
         */
        let initEvents = () => {
            s.opts.elm.feedback.faq.children("strong").on("click", (e) => { // faq toggle
                e.preventDefault();
                $(e.currentTarget).next("p").toggleClass(s.opts.classes.visible);
            });

            s.opts.elm.feedback.faq.children("p > a").on("click", (e) => { // handle links inside the faq answers
                e.preventDefault();
                let type = $(e.currentTarget).parent("p").attr(s.opts.attr.type);

                if (type === "usage") {
                    window.open(chrome.extension.getURL("html/intro.html") + "?skip=1", '_blank');
                }
            });

            s.opts.elm.feedback.send.on("click", (e) => { // submit feedback form
                e.preventDefault();
                sendFeedback();
            });
        };

        /**
         * Checks the content of the feedback fields and sends the content via ajax if they are filled properly
         */
        let sendFeedback = () => {
            let messageText = s.opts.elm.textarea.feedbackMsg[0].value;
            let emailText = s.opts.elm.field.feedbackEmail[0].value;
            let isEmailValid = emailText.length > 0 && /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(emailText);
            let isMessageValid = messageText.length > 0;

            if (isEmailValid && isMessageValid) {
                let loadStartTime = +new Date();
                let loader = s.helper.template.loading().appendTo(s.opts.elm.body);
                s.opts.elm.body.addClass(s.opts.classes.loading);
                let infos = null;

                $.xhr(s.opts.ajax.feedback, {
                    method: "POST",
                    data: {
                        email: emailText,
                        msg: messageText,
                        version: s.opts.manifest.version,
                        ua: navigator.userAgent,
                        lang: s.helper.i18n.getLanguage(),
                        config: s.helper.importExport.getExportConfig()
                    }
                }).then((xhr) => {
                    infos = JSON.parse(xhr.responseText);
                    return $.delay(Math.max(0, 1000 - (+new Date() - loadStartTime))); // load at least 1s
                }).then(() => {
                    s.opts.elm.body.removeClass(s.opts.classes.loading);
                    loader.remove();

                    if (infos && infos.success && infos.success === true) { // successfully submitted -> show message and clear form
                        s.opts.elm.textarea.feedbackMsg[0].value = "";
                        s.opts.elm.field.feedbackEmail[0].value = "";
                        s.showSuccessMessage("feedback_sent_message");
                    } else { // not submitted -> raise error
                        $.delay().then(() => {
                            alert(s.helper.i18n.get("settings_feedback_send_failed"));
                        });
                    }
                });
            } else if (!isEmailValid) {
                s.opts.elm.field.feedbackEmail.addClass(s.opts.classes.error);
            } else if (!isMessageValid) {
                s.opts.elm.textarea.feedbackMsg.addClass(s.opts.classes.error);
            }

            $.delay(700).then(() => {
                $("." + s.opts.classes.error).removeClass(s.opts.classes.error);
            });
        };
    };

})(jsu);