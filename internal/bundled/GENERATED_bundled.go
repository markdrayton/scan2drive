package bundled

// Table of contents
var assets = map[string][]byte{
	"assets/scan2drive.js": assets_0,
}
var assets_0 = []byte("// vim:ts=4:sw=4:et\n// Documentation references:\n// https://developers.google.com/identity/sign-in/web/reference\n// https://developers.google.com/picker/docs/reference\n\nfunction httpErrorToToast(jqXHR, prefix) {\n    var summary = 'HTTP ' + jqXHR.status + ': ' + jqXHR.responseText;\n    Materialize.toast(prefix + ': ' + summary, 5000, 'red');\n    console.log('error: prefix', prefix, ', summary', summary);\n}\n\n// start is called once the Google APIs were loaded\nfunction start() {\n    console.log('start');\n\n    gapi.load('auth2', function() {\n	gapi.client.load('plus', 'v1').then(function() {\n            var auth2 = gapi.auth2.init({\n		client_id: clientID,\n		// The “profile” and “email” scope are always requested.\n		scope: 'https://www.googleapis.com/auth/drive',\n            });\n            auth2.then(function() {\n		user = auth2.currentUser.get();\n		var sub = $('#user-name').attr('data-sub');\n		// If sub does not match the user id, we are logged in in the\n		// browser, but not on the server side (e.g. because sessions were\n		// deleted).\n		if (auth2.isSignedIn.get() && user.getId() === sub) {\n		    console.log('logged in');\n                    $('#user-avatar').attr('src', user.getBasicProfile().getImageUrl());\n                    $('#user-name').text(user.getBasicProfile().getName());\n                    $('.fixed-action-btn').show();\n                    $('#signin').hide();\n                    $('#signout').show();\n                    $('#settings-button').show();\n                    // TODO: open settings button in case drive folder is not configured\n\n		    // Resolve user ids into names and thumbnails for the people dialog\n		    $('div.user').each(function(idx, el) {\n			var sub = $(el).data('sub');\n			var req = gapi.client.plus.people.get({'userId': sub});\n			req.execute(function(result) {\n			    var nick = result.displayName;\n			    var thumb = result.image.url;\n			    $('div.user[data-sub=\"' + sub + '\"] img').attr('src', thumb);\n			    $('div.user[data-sub=\"' + sub + '\"] span.user-nick').text(nick);\n			});\n		    });\n		} else {\n		    console.log('auth2 loaded, but user not logged in');\n		    gapi.signin2.render('my-signin2', {\n			'scope': 'profile email',\n			'width': 240,\n			'height': 50,\n			'longtitle': true,\n			'theme': 'dark',\n			'onsuccess': function(){ console.log('success'); },\n			'onfailure': function() { console.log('failure'); }\n		    });\n		}\n            }, function(err) {\n		var errorp = $('#error p');\n		errorp.text('Error ' + err.error + ': ' + err.details);\n		console.log('OAuth2 error', err);\n            });\n	});\n    });\n\n    gapi.load('picker', function() {});\n\n    $('#signinButton').click(function() {\n	var auth2 = gapi.auth2.getAuthInstance();\n        auth2.grantOfflineAccess({'redirect_uri': 'postmessage'}).then(signInCallback);\n    });\n\n    $('#signout').click(function(ev) {\n        var auth2 = gapi.auth2.getAuthInstance();\n        auth2.signOut().then(function() {\n            $.ajax({\n                type: 'POST',\n                url: '/signout',\n                success: function(result) {\n                    // Reload the page.\n                    window.location.href = window.location.origin;\n                },\n                // TODO: error handling (deleting file failed, e.g. because of readonly file system)\n            });\n        });\n        ev.preventDefault();\n    });\n\n    // TODO: #signin keypress\n    $('#signin').click(function(ev) {\n	var auth2 = gapi.auth2.getAuthInstance();\n        auth2.grantOfflineAccess({'redirect_uri': 'postmessage'}).then(signInCallback);\n        ev.preventDefault();\n    });\n\n    $('#select-drive-folder').click(function() {\n        createPicker();\n    });\n\n    $('div.user').click(function() {\n	var sub = $(this).data('sub');\n	// TODO: show progress spinner\n	$.ajax({\n            type: 'POST',\n            url: '/storedefaultuser',\n            contentType: 'application/json',\n            success: function(data, textStatus, jqXHR) {\n		$('div.user').removeClass('default-user');\n		$('div.user[data-sub=\"' + sub + '\"]').addClass('default-user');\n		$('#people-dialog').modal('close');\n            },\n            error: function(jqXHR, textStatus, errorThrown) {\n		httpErrorToToast(jqXHR, 'storing default user failed');\n            },\n            processData: false,\n            data: JSON.stringify({\n		DefaultSub: sub,\n            }),\n	});\n    });\n}\n\nfunction pollScan(name) {\n    var user = gapi.auth2.getAuthInstance().currentUser.get();\n    $.ajax({\n        type: 'POST',\n        url: '/scanstatus',\n        contentType: 'application/json',\n        success: function(data, textStatus, jqXHR) {\n            $('#scan-progress-status').text(data.Status);\n            console.log('result', data, 'textStatus', textStatus, 'jqXHR', jqXHR);\n            if (jqXHR.status !== 200) {\n                // TODO: show error message\n                return;\n            }\n            if (data.Done) {\n                $('#scan-dialog paper-spinner-lite').attr('active', null); // TODO\n                $('.fixed-action-btn i').text('scanner');\n                $('.fixed-action-btn a').removeClass('disabled');\n                $('#scan-dialog').off('iron-overlay-canceled');\n                var sub = user.getBasicProfile().getId();\n                $('#scan-dialog .scan-thumb').css('background', 'url(\"scans_dir/' + sub + '/' + name + '/thumb.png\")').css('background-size', 'cover');\n            } else {\n                setTimeout(function() { pollScan(name); }, 500);\n            }\n        },\n        error: function(jqXHR, textStatus, errorThrown) {\n            if (jqXHR.status === 404) {\n                // Scan was not yet found because the directory rescan isn’t done.\n                // Retry in a little while.\n                setTimeout(function() { pollScan(name); }, 500);\n            } else {\n                $('#scan-progress-status').text('Error: ' + errorThrown);\n                setTimeout(function() { pollScan(name); }, 500);\n            }\n        },\n        processData: false,\n        data: JSON.stringify({'Name':name}),\n    });\n}\n\nfunction renameScan(name, newSuffix) {\n    var newName = name + '-' + newSuffix;\n\n    $.ajax({\n        type: 'POST',\n        url: '/renamescan',\n        contentType: 'application/json',\n        success: function(data, textStatus, jqXHR) {\n            $('#scan-form paper-input iron-icon').show();\n        },\n        error: function(jqXHR, textStatus, errorThrown) {\n            httpErrorToToast(jqXHR, 'renaming scan failed');\n        },\n        processData: false,\n        data: JSON.stringify({\n            Name: name,\n            NewName: newName,\n        }),\n    });\n}\n\nfunction scan() {\n    // Only one scan can be in progress at a time.\n    $('.fixed-action-btn i').text('hourglass_empty');\n    $('.fixed-action-btn a').addClass('disabled');\n    $('#scan-dialog').modal('open');\n\n    $.ajax({\n        type: 'POST',\n        url: '/startscan',\n        success: function(data, textStatus, jqXHR) {\n            $('#scan-dialog paper-input[name=\"name\"] div[prefix]').text(data.Name + '-');\n            var renameButton = $('#scan-form paper-button');\n            renameButton.click(function(ev) {\n                renameScan(data.Name, $('#scan-form paper-input').val());\n            });\n            pollScan(data.Name);\n        },\n        error: function(jqXHR, textStatus, errorThrown) {\n            $('#scan-dialog').modal('close');\n            $('.fixed-action-btn i').text('scanner');\n            $('.fixed-action-btn a').removeClass('disabled');\n            httpErrorToToast(jqXHR, 'scanning failed');\n        },\n    });\n}\n\n// callback has “loaded”, “cancel” and “picked”\nfunction pickerCallback(data) {\n    console.log('picker callback', data);\n    if (data.action !== google.picker.Action.PICKED) {\n        return;\n    }\n    if (data.docs.length !== 1) {\n        // TODO: error handling\n        return;\n    }\n    var picked = data.docs[0];\n    // TODO: show a spinner\n    $.ajax({\n        type: 'POST',\n        url: '/storedrivefolder',\n        contentType: 'application/json',\n        success: function(data, textStatus, jqXHR) {\n            $('#drivefolder').val(picked.name);\n        },\n        error: function(jqXHR, textStatus, errorThrown) {\n            httpErrorToToast(jqXHR, 'storing drive folder failed');\n        },\n        data: JSON.stringify({\n            'Id': picked.id,\n            'IconUrl': picked.iconUrl,\n            'Url': picked.url,\n            'Name': picked.name,\n        }),\n    });\n}\n\nfunction createPicker() {\n    var user = gapi.auth2.getAuthInstance().currentUser.get();\n\n    if (!user) {\n        // The picker requires an OAuth token.\n        return;\n    }\n\n    var docsView = new google.picker.DocsView()\n        .setIncludeFolders(true)\n        .setMimeTypes('application/vnd.google-apps.folder')\n        .setMode(google.picker.DocsViewMode.LIST)\n        .setSelectFolderEnabled(true);\n\n    var picker = new google.picker.PickerBuilder()\n        .addView(docsView)\n        .setCallback(pickerCallback)\n        .setOAuthToken(user.getAuthResponse().access_token)\n        .build();\n    picker.setVisible(true);\n}\n\nfunction signInCallback(authResult) {\n    if (authResult['code']) {\n        // TODO: progress indicator, writing to disk and examining scans could take a while.\n        $.ajax({\n            type: 'POST',\n            url: '/oauth',\n            contentType: 'application/octet-stream; charset=utf-8',\n            success: function(result) {\n                // Reload the page.\n                window.location.href = window.location.origin;\n            },\n            error: function(jqXHR, textStatus, errorThrown) {\n		if (jqXHR.status == 500) {\n                    $('#error p').text('OAuth error: ' + jqXHR.responseText + '. Try revoking access on https://security.google.com/settings/u/0/security/permissions, then retry');\n		} else {\n                    $('#error p').text('Unknown OAuth error: ' + errorThrown);\n                }\n            },\n            processData: false,\n            data: authResult['code'],\n        });\n    } else {\n        console.log('sth went wrong :|', authResult);\n        // TODO: trigger logout, without server-side auth we are screwed\n    }\n}\n")
