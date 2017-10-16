($ => {
    "use strict";

    window.settings = function () {

        /*
         * ################################
         * PUBLIC
         * ################################
         */

        this.opts = {
            classes: {
                page: {
                    darkMode: "dark"
                },
                tabs: {
                    content: "tab",
                    active: "active"
                },
                color: {
                    field: "color",
                    mask: "colorMask",
                    suggestion: "suggestion"
                },
                radio: {
                    wrapper: "radioWrapper"
                },
                range: {
                    inactive: "inactive"
                },
                newtab: {
                    hideable: "hideable"
                },
                translation: {
                    select: "languageSelect",
                    category: "category",
                    edit: "edit",
                    progress: "progress",
                    mark: "mark",
                    requiredInfo: "requiredInfo",
                    amountInfo: "amountInfo",
                    empty: "empty",
                    back: "back",
                    hover: "hover",
                    goto: "goto"
                },
                checkbox: {
                    box: "checkbox",
                    active: "active",
                    clicked: "clicked",
                    focus: "focus"
                },
                hidden: "hidden",
                success: "success",
                error: "error",
                building: "building",
                initLoading: "initLoading",
                loading: "loading",
                revert: "revert",
                visible: "visible",
                small: "small",
                desc: "desc",
                box: "box",
                dialog: "dialog",
                boxWrapper: "boxWrapper",
                contentBox: "contentBox",
                action: "action",
                incomplete: "incomplete"
            },
            attr: {
                type: "data-type",
                appearance: "data-appearance",
                name: "data-name",
                i18n: "data-i18n",
                value: "data-value",
                success: "data-successtext",
                style: "data-style",
                hideOnFalse: "data-hideOnFalse",
                buttons: {
                    save: "data-save",
                    restore: "data-restore"
                },
                range: {
                    min: "data-min",
                    max: "data-max",
                    step: "data-step",
                    unit: "data-unit",
                    infinity: "data-infinity"
                },
                color: {
                    alpha: "data-alpha",
                    style: "data-color",
                    suggestions: "data-suggestions"
                },
                field: {
                    placeholder: "data-placeholder"
                },
                translation: {
                    releaseStatus: "data-status",
                    language: "data-lang"
                }
            },
            elm: {
                body: $("body"),
                title: $("head > title"),
                aside: $("body > section#wrapper > aside"),
                content: $("body > section#wrapper > main"),
                header: $("body > header"),
                headline: $("body > header > h1"),
                buttons: {
                    save: $("body > header > menu > button.save"),
                    restore: $("body > header > menu > button.restore"),
                    'import': $("body a.import > input[type='file']"),
                    'export': $("body a.export"),
                },
                appearance: {
                    content: $("div.tab[data-name='appearance']"),
                },
                newtab: {
                    content: $("div.tab[data-name='newtab']"),
                },
                feedback: {
                    wrapper: $("div.tab[data-name='feedback']"),
                    form: $("section.form"),
                    send: $("section.form button[type='submit']"),
                    faq: $("div.faq")
                },
                translation: {
                    wrapper: $("div.tab[data-name='language'] > div[data-name='translate']"),
                    goto: $("div.tab[data-name='language'] > div[data-name='general'] button[type='submit']"),
                    overview: $("div.tab[data-name='language'] > div[data-name='translate'] > div.overview"),
                    langvars: $("div.tab[data-name='language'] > div[data-name='translate'] > div.langvars"),
                    unavailable: $("div.tab[data-name='language'] > div[data-name='translate'] > div.unavailable")
                },
                keyboardShortcutInfo: $("p.shortcutInfo"),
                formElement: $("div.formElement"),
                support: {
                    donateButton: $("div.tab[data-name='support'] button[type='submit']")
                },
                preview: {},
                checkbox: {},
                range: {},
                select: {},
                color: {},
                textarea: {},
                field: {},
                radio: {}
            },
            events: {
                pageChanged: "blockbyte-bs-pageChanged"
            },
            ajax: {
                feedback: "https://extensions.blockbyte.de/ajax/feedback",
                translation: {
                    info: "https://extensions.blockbyte.de/ajax/bs/i18n/info",
                    langvars: "https://extensions.blockbyte.de/ajax/bs/i18n/langvars",
                    submit: "https://extensions.blockbyte.de/ajax/bs/i18n/submit"
                }
            },
            donateLink: "https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=2VW2UADL99YEL",
            manifest: chrome.runtime.getManifest()
        };

        this.serviceAvailable = true;
        let restoreTypes = ["sidebar", "appearance", "newtab"];

        /**
         * Constructor
         */
        this.run = () => {
            initHelpers();
            let loader = {
                body: this.helper.template.loading().appendTo(this.opts.elm.body)
            };
            this.opts.elm.body.addClass(this.opts.classes.initLoading);

            this.helper.model.init().then(() => {
                return this.helper.i18n.init();
            }).then(() => {
                this.helper.font.init('default');
                this.helper.stylesheet.init();
                this.helper.stylesheet.addStylesheets(["settings"], $(document));
                initHeader();

                return this.helper.form.init();
            }).then(() => {
                this.opts.elm.body.removeClass(this.opts.classes.building);

                this.helper.i18n.parseHtml(document);
                this.opts.elm.title.text(this.opts.elm.title.text() + " - " + this.helper.i18n.get("extension_name"));

                ["translation", "feedback"].forEach((name) => {
                    loader[name] = this.helper.template.loading().appendTo(this.opts.elm[name].wrapper);
                    this.opts.elm[name].wrapper.addClass(this.opts.classes.loading);
                });

                return Promise.all([
                    this.helper.menu.init(),
                    this.helper.sidebar.init(),
                    this.helper.appearance.init(),
                    this.helper.newtab.init(),
                    this.helper.support.init(),
                    this.helper.importExport.init(),
                ]);
            }).then(() => { // initialise events and remove loading mask
                initEvents();

                loader.body.remove();
                this.opts.elm.body.removeClass(this.opts.classes.initLoading);
                this.helper.model.call("trackPageView", {page: "/settings"});

                return this.helper.model.call("websiteStatus");
            }).then((opts) => { // if website is available, feedback form and translation overview can be used
                this.serviceAvailable = opts.status === "available";
                this.helper.feedback.init();
                this.helper.translation.init();

                ["translation", "feedback"].forEach((name) => {
                    loader[name].remove();
                    this.opts.elm[name].wrapper.removeClass(this.opts.classes.loading);
                });
            });
        };

        /**
         * Shows the given success message for 1.5s
         *
         * @param {string} i18nStr
         */
        this.showSuccessMessage = (i18nStr) => {
            this.opts.elm.body.attr(this.opts.attr.success, this.helper.i18n.get("settings_" + i18nStr));
            this.opts.elm.body.addClass(this.opts.classes.success);

            $.delay(1500).then(() => {
                this.opts.elm.body.removeClass(this.opts.classes.success);
            });
        };

        /*
         * ################################
         * PRIVATE
         * ################################
         */

        /**
         * Initialises the helper objects
         */
        let initHelpers = () => {
            this.helper = {
                model: new window.ModelHelper(this),
                checkbox: new window.CheckboxHelper(this),
                template: new window.TemplateHelper(this),
                i18n: new window.I18nHelper(this),
                font: new window.FontHelper(this),
                stylesheet: new window.StylesheetHelper(this),
                translation: new window.TranslationHelper(this),
                menu: new window.MenuHelper(this),
                form: new window.FormHelper(this),
                sidebar: new window.SidebarHelper(this),
                newtab: new window.NewtabHelper(this),
                appearance: new window.AppearanceHelper(this),
                feedback: new window.FeedbackHelper(this),
                importExport: new window.ImportExportHelper(this),
                support: new window.SupportHelper(this)
            };
        };

        let initHeader = () => {
            this.helper.template.svgByName("icon-settings").then((svg) => {
                this.opts.elm.header.prepend(svg);
            });
        };

        /**
         * Initialises the eventhandlers
         *
         * @returns {Promise}
         */
        let initEvents = async () => {
            $(document).on("click", () => {
                $("div." + this.opts.classes.dialog).removeClass(this.opts.classes.visible);
            });

            this.opts.elm.header.on("click", "div." + this.opts.classes.dialog, (e) => {
                e.stopPropagation();
            });

            this.opts.elm.header.on("click", "div." + this.opts.classes.dialog + " > a", (e) => {
                e.preventDefault();
                e.stopPropagation();

                let type = $(e.currentTarget).parent("div").attr(this.opts.attr.type);

                if (restoreTypes.indexOf(type) !== -1) {
                    let language = this.helper.model.getData("b/language");

                    chrome.storage.sync.remove([type], () => {
                        if (type === "sidebar") { // don't reset user language
                            chrome.storage.sync.set({behaviour: {language: language}});
                        }

                        this.showSuccessMessage("restored_message");
                        this.helper.model.call("reloadIcon");
                        $("div." + this.opts.classes.dialog).removeClass(this.opts.classes.visible);

                        $.delay(1500).then(() => {
                            this.helper.model.call("reinitialize");
                            location.reload(true);
                        });
                    });
                }
            });

            this.opts.elm.buttons.save.on("click", (e) => { // save button
                e.preventDefault();
                let path = this.helper.menu.getPath();

                switch (path[0]) {
                    case "sidebar": {
                        this.helper.sidebar.save();
                        break;
                    }
                    case "appearance": {
                        this.helper.appearance.save();
                        break;
                    }
                    case "newtab": {
                        this.helper.newtab.save();
                        break;
                    }
                    case "language": {
                        if (path[1] === "translate") {
                            this.helper.translation.submit();
                        } else {
                            this.helper.sidebar.saveLanguage().then(() => {
                                return $.delay(1500);
                            }).then(() => {
                                location.reload(true);
                            });
                        }
                        break;
                    }
                }
            });

            this.opts.elm.buttons.restore.on("click", (e) => {
                e.preventDefault();
                let path = this.helper.menu.getPath();

                if (restoreTypes.indexOf(path[0]) !== -1) {
                    $("div." + this.opts.classes.dialog).remove();
                    let dialog = $("<div />")
                        .attr(this.opts.attr.type, path[0])
                        .addClass(this.opts.classes.dialog)
                        .append("<p>" + this.helper.i18n.get("settings_restore_confirm") + "</p>")
                        .append("<span>" + this.helper.i18n.get("settings_menu_" + path[0]) + "</span>")
                        .append("<br />")
                        .append("<a>" + this.helper.i18n.get("settings_restore") + "</a>")
                        .css("right", this.opts.elm.header.css("padding-right"))
                        .appendTo(this.opts.elm.header);

                    $.delay().then(() => {
                        dialog.addClass(this.opts.classes.visible);
                    });
                }
            });
        };
    };

    new window.settings().run();
})(jsu);