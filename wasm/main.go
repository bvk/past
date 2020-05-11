// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bvk/past/msg"
	"golang.org/x/xerrors"
	"honnef.co/go/js/dom/v2"
)

func main() {
	if err := doMain(); err != nil {
		log.Fatal(err)
	}
}

func doMain() error {
	window := dom.GetWindow()
	location := window.Location()
	protocol := location.Protocol()

	var backend Backend
	if strings.EqualFold(protocol, "chrome-extension:") {
		log.Println("Using chrome extension backend")
		b, err := NewExtensionBackend()
		if err != nil {
			return xerrors.Errorf("could not create chrome-extension backend: %w", err)
		}
		backend = b
	} else {
		log.Println("Using server backend with ajax")
		b, err := NewServerBackend(location.Href())
		if err != nil {
			return xerrors.Errorf("could not create server backend: %w", err)
		}
		backend = b
	}

	log.Println("Location", location)
	log.Println("Location.Protocol", location.Protocol())
	log.Println("Location.Host", location.Host())
	log.Println("Location.Hostname", location.Hostname())
	log.Println("Location.Href", location.Href())
	log.Println("Location.Origin", location.Origin())

	ctl, err := NewController(backend)
	if err != nil {
		return xerrors.Errorf("could not create controller: %w", err)
	}

	// // Show all pages side-by-side.
	// pages, err := allPages(ctl)
	// if err != nil {
	// 	return xerrors.Errorf("could not create all page instances: %w", err)
	// }
	// if err := ctl.ShowAllPages(pages); err != nil {
	// 	return xerrors.Errorf("could not show all pages: %w", err)
	// }

	resp, err := ctl.CheckStatus(&msg.CheckStatusRequest{ListFiles: true})
	if err != nil {
		return xerrors.Errorf("could not check status: %w", err)
	}
	listFiles := resp.CheckStatus.ListFiles

	// Add context menu entries if backend supports context menus. Context menus
	// are supported if we are running as a chrome/chromium extension.
	if cm, ok := backend.(ContextMenuSupport); ok && listFiles != nil {
		ids, err := createOrUpdateMenus(cm, listFiles.Files, ctl.MenuIDs())
		if err := ctl.UpdateMenuIDs(ids); err != nil {
			return xerrors.Errorf("could not update menu ids: %w", err)
		}
		if err != nil {
			return xerrors.Errorf("could not add context menus: %w", err)
		}
	}

	if listFiles == nil {
		page, err := NewSettingsPage(ctl, resp)
		if err != nil {
			return xerrors.Errorf("could not create settings page: %w", err)
		}
		ctl.ShowPage(page)
	} else {
		page, err := NewSearchPage(ctl, listFiles, ctl.UseCountMap())
		if err != nil {
			return xerrors.Errorf("could not create list page: %w", err)
		}
		ctl.ShowPage(page)
	}

	<-make(chan struct{})
	return nil
}

func createOrUpdateMenus(cm ContextMenuSupport, files, oldMenuIDs []string) ([]string, error) {
	contexts := []string{"editable"}

	oldIDMap := make(map[string]struct{})
	for _, id := range oldMenuIDs {
		oldIDMap[id] = struct{}{}
	}

	var newIDs []string
	for _, file := range files {
		file := file
		if _, ok := oldIDMap[file]; ok {
			newIDs = append(newIDs, file)
			continue
		}
		patterns, err := siteURLPatterns(file)
		if err != nil {
			continue
		}
		if err := cm.AddContextMenu(file, file, contexts, patterns); err != nil {
			return newIDs, xerrors.Errorf("could not create or update context menu for %q: %w", file, err)
		}
		newIDs = append(newIDs, file)
		log.Printf("added context menu %q to pages %v", file, patterns)
	}

	// Also, remove menu items that doesn't exist anymore.
	fileMap := make(map[string]struct{})
	for _, file := range files {
		fileMap[file] = struct{}{}
	}
	for id := range oldIDMap {
		if _, ok := fileMap[id]; !ok {
			if err := cm.RemoveContextMenu(id); err != nil {
				log.Printf("warning: could not remove stale menu item %q (ignored)", id)
			} else {
				log.Printf("removed stale context menu %q", id)
			}
		}
	}

	return newIDs, nil
}

func siteURLPatterns(filename string) ([]string, error) {
	dir := filepath.Dir(filename)
	if len(dir) == 0 || dir == "." {
		return nil, xerrors.Errorf("could not determine sitename component: %w", os.ErrInvalid)
	}
	// No subdirectories are allowed.
	if strings.ContainsRune(dir, filepath.Separator) {
		return nil, xerrors.Errorf("directory cannot contain subdirectories: %w", os.ErrInvalid)
	}
	words := strings.Split(dir, ".")
	if len(words) <= 1 {
		return nil, xerrors.Errorf("directory %q is not a valid sitename: %w", dir, os.ErrInvalid)
	}
	var patterns []string
	for ii := 0; ii < len(words)-1; ii++ {
		pattern := fmt.Sprintf("https://*.%s/*", strings.Join(words[ii:], "."))
		patterns = append(patterns, pattern)
	}
	return patterns, nil
}

func allPages(ctl *Controller) ([]Page, error) {
	response := new(msg.Response)
	reader := strings.NewReader(ExampleCheckResponse)
	if err := json.NewDecoder(reader).Decode(response); err != nil {
		return nil, err
	}
	settings, err := NewSettingsPage(ctl, response)
	if err != nil {
		return nil, err
	}
	addkey, err := NewAddKeyPage(ctl)
	if err != nil {
		return nil, err
	}
	keyring, err := NewKeyringPage(ctl, response.CheckStatus)
	if err != nil {
		return nil, err
	}
	key, err := NewKeyPage(ctl, response.CheckStatus.LocalKeys[0])
	if err != nil {
		return nil, err
	}
	initpast, err := NewInitPastPage(ctl, response.CheckStatus)
	if err != nil {
		return nil, err
	}
	past, err := NewPastPage(ctl, response.ScanStore)
	if err != nil {
		return nil, err
	}
	initremote, err := NewInitRemotePage(ctl)
	if err != nil {
		return nil, err
	}
	remote, err := NewRemotePage(ctl, response.SyncRemote)
	if err != nil {
		return nil, err
	}
	search, err := NewSearchPage(ctl, response.ListFiles, make(map[string]int))
	if err != nil {
		return nil, err
	}
	view, err := NewViewPage(ctl, response.ViewFile)
	if err != nil {
		return nil, err
	}
	edit, err := NewEditPage(ctl, response.ViewFile)
	if err != nil {
		return nil, err
	}
	pages := []Page{
		settings, addkey, keyring, key, initpast, past, initremote, remote, search, view, edit,
	}
	return pages, nil
}

const ExampleCheckResponse = `
{
  "status": "",
  "check_status": {
    "gpg_path": "/usr/bin/gpg",
    "git_path": "/usr/bin/git",
    "local_keys": [
      {
        "key_id": "621AD099AAFE93B9",
        "key_length": 4096,
        "key_fingerprint": "7134C24A8899FFCB9BD43719621AD099AAFE93B9",
        "user_name": "BVK Chaitanya",
        "user_email": "bvk@steam.zion.sh",
        "is_subkey": true,
        "is_trusted": true,
        "is_expired": false,
        "can_encrypt": true,
        "can_decrypt": true,
        "days_to_expire": 307
      },
      {
        "key_id": "AAEC846975159E8E",
        "key_length": 4096,
        "key_fingerprint": "15CBE88E5A1C3970FD9D77EFAAEC846975159E8E",
        "user_name": "Blue Rose",
        "user_email": "blue@rose.com",
        "is_subkey": true,
        "is_trusted": true,
        "is_expired": false,
        "can_encrypt": true,
        "can_decrypt": true,
        "days_to_expire": 362
      },
      {
        "key_id": "062747E8AFCF72EF",
        "key_length": 1024,
        "key_fingerprint": "1F9639EF47D0544D31B085CC062747E8AFCF72EF",
        "user_name": "Water Works",
        "user_email": "water@works.com",
        "is_subkey": true,
        "is_trusted": true,
        "is_expired": false,
        "can_encrypt": true,
        "can_decrypt": true,
        "days_to_expire": 0
      },
      {
        "key_id": "68A52BDB9C091D3A",
        "key_length": 4096,
        "key_fingerprint": "5D1AD7D6B1B3E479B2D8416F68A52BDB9C091D3A",
        "user_name": "Rolls Royce",
        "user_email": "rolls@royce.com",
        "is_subkey": true,
        "is_trusted": true,
        "is_expired": false,
        "can_encrypt": true,
        "can_decrypt": true,
        "days_to_expire": 0
      },
      {
        "key_id": "0E53F1CC7768A695",
        "key_length": 4096,
        "key_fingerprint": "D4AF24D6D316116F405795540E53F1CC7768A695",
        "user_name": "Red Wedding",
        "user_email": "red@wedding.com",
        "is_subkey": true,
        "is_trusted": true,
        "is_expired": false,
        "can_encrypt": true,
        "can_decrypt": true,
        "days_to_expire": 0
      }
    ],
    "remote_keys": [
      {
        "key_id": "986C05EF92889B3A",
        "key_length": 2048,
        "key_fingerprint": "773CB05C940D0EE4762BA4FE986C05EF92889B3A",
        "user_name": "BVK Chaitanya",
        "user_email": "bvkchaitanya@gmail.com",
        "is_subkey": true,
        "is_trusted": true,
        "is_expired": false,
        "can_encrypt": true,
        "can_decrypt": false,
        "days_to_expire": 0
      },
      {
        "key_id": "B59383D48832BFE8",
        "key_length": 4096,
        "key_fingerprint": "D7655F3528215B066B4F8DB1B59383D48832BFE8",
        "user_name": "Bitcoin ABC Security",
        "user_email": "security@bitcoinabc.org",
        "is_subkey": true,
        "is_trusted": false,
        "is_expired": false,
        "can_encrypt": true,
        "can_decrypt": false,
        "days_to_expire": 276
      }
    ],
    "expired_keys": null,
    "password_store_keys": [
      "2E42F5A54810D7BCE938D3A6DD52E2A86AFFBFCC",
      "15CBE88E5A1C3970FD9D77EFAAEC846975159E8E",
      "D4AF24D6D316116F405795540E53F1CC7768A695"
    ],
    "remote": ""
  },
  "create_key": null,
  "import_key": null,
  "edit_key": null,
  "export_key": null,
  "delete_key": null,
  "create_repo": null,
  "import_repo": null,
  "add_remote": null,
  "sync_remote": {
    "head": {
      "commit":"628dac8c60a4a90dfa43b40888140ba47cedc06e",
      "author":"BVK Chaitanya <bvk@steam.zion.sh>",
      "author_date":"2014-05-16T08:28:06.801064-04:00",
      "title":"Updated password file foo.com.gpg"
    },
    "remote": {
      "commit":"628dac8c60a4a90dfa43b40888140ba47cedc06e",
      "author":"BVK Chaitanya <bvk@steam.zion.sh>",
      "author_date":"2014-05-16T08:28:06.801064-04:00",
      "title":"Updated password file foo.com.gpg"
    },
    "newer_commit":""
  },
  "scan_store": {
    "num_files": 105,
    "key_map": {
      "15CBE88E5A1C3970FD9D77EFAAEC846975159E8E": {
        "key_id": "AAEC846975159E8E",
        "key_length": 4096,
        "key_fingerprint": "15CBE88E5A1C3970FD9D77EFAAEC846975159E8E",
        "user_name": "Blue Rose",
        "user_email": "blue@rose.com",
        "is_subkey": true,
        "is_trusted": true,
        "is_expired": false,
        "can_encrypt": true,
        "can_decrypt": true,
        "days_to_expire": 362
      },
      "773CB05C940D0EE4762BA4FE986C05EF92889B3A": {
        "key_id": "986C05EF92889B3A",
        "key_length": 2048,
        "key_fingerprint": "773CB05C940D0EE4762BA4FE986C05EF92889B3A",
        "user_name": "BVK Chaitanya",
        "user_email": "bvkchaitanya@gmail.com",
        "is_subkey": true,
        "is_trusted": true,
        "is_expired": false,
        "can_encrypt": true,
        "can_decrypt": false,
        "days_to_expire": 0
      },
      "D4AF24D6D316116F405795540E53F1CC7768A695": {
        "key_id": "0E53F1CC7768A695",
        "key_length": 4096,
        "key_fingerprint": "D4AF24D6D316116F405795540E53F1CC7768A695",
        "user_name": "Red Wedding",
        "user_email": "red@wedding.com",
        "is_subkey": true,
        "is_trusted": true,
        "is_expired": false,
        "can_encrypt": true,
        "can_decrypt": true,
        "days_to_expire": 0
      }
    },
    "unused_key_map": {
      "1F9639EF47D0544D31B085CC062747E8AFCF72EF": {
        "key_id": "062747E8AFCF72EF",
        "key_length": 1024,
        "key_fingerprint": "1F9639EF47D0544D31B085CC062747E8AFCF72EF",
        "user_name": "Water Works",
        "user_email": "water@works.com",
        "is_subkey": true,
        "is_trusted": true,
        "is_expired": false,
        "can_encrypt": true,
        "can_decrypt": true,
        "days_to_expire": 0
      },
      "5D1AD7D6B1B3E479B2D8416F68A52BDB9C091D3A": {
        "key_id": "68A52BDB9C091D3A",
        "key_length": 4096,
        "key_fingerprint": "5D1AD7D6B1B3E479B2D8416F68A52BDB9C091D3A",
        "user_name": "Rolls Royce",
        "user_email": "rolls@royce.com",
        "is_subkey": true,
        "is_trusted": true,
        "is_expired": false,
        "can_encrypt": true,
        "can_decrypt": true,
        "days_to_expire": 0
      },
      "7134C24A8899FFCB9BD43719621AD099AAFE93B9": {
        "key_id": "621AD099AAFE93B9",
        "key_length": 4096,
        "key_fingerprint": "7134C24A8899FFCB9BD43719621AD099AAFE93B9",
        "user_name": "BVK Chaitanya",
        "user_email": "bvk@steam.zion.sh",
        "is_subkey": true,
        "is_trusted": true,
        "is_expired": false,
        "can_encrypt": true,
        "can_decrypt": true,
        "days_to_expire": 306
      },
      "D7655F3528215B066B4F8DB1B59383D48832BFE8": {
        "key_id": "B59383D48832BFE8",
        "key_length": 4096,
        "key_fingerprint": "D7655F3528215B066B4F8DB1B59383D48832BFE8",
        "user_name": "Bitcoin ABC Security",
        "user_email": "security@bitcoinabc.org",
        "is_subkey": true,
        "is_trusted": false,
        "is_expired": false,
        "can_encrypt": true,
        "can_decrypt": false,
        "days_to_expire": 276
      }
    },
    "key_file_count_map": {
      "15CBE88E5A1C3970FD9D77EFAAEC846975159E8E": 105,
      "773CB05C940D0EE4762BA4FE986C05EF92889B3A": 105,
      "D4AF24D6D316116F405795540E53F1CC7768A695": 105
    },
    "missing_key_file_count_map": {}
  },
  "add_recipient": null,
  "remove_recipient": null,
  "add_file": null,
  "edit_file": null,
  "list_files": {
    "files": [
      "401k.moneyintel.com/cbayapuneni@cloudsimple.com",
      "accounts.google.com/bvk.other@gmail.com",
      "accounts.intuit.com/bvkchaitanya@gmail.com",
      "accounts.intuit.com/i_chaitu@yahoo.com",
      "api.blockpress.com/bvk",
      "app.pagerduty.com/cbayapuneni@cloudsimple.com",
      "app.swaggerhub.com/bvkchaitanya",
      "applications.wes.org/bvk.other@gmail.com",
      "applications.wes.org/i_chaitu@yahoo.com",
      "auth.hitbtc.com/bvkother@gmail.com",
      "auth.hitbtc.com/i_chaitu@yahoo.com",
      "bchgang.com/bvk",
      "bitbucket.org/bvkchaitanya@gmail.com",
      "bitcoin.tax/bvkother@gmail.com",
      "bitflip.li/bvkother@gmail.com",
      "bitgrail.com/bvkother@gmail.com",
      "calendar.google.com/cbayapuneni@cloudsimple.com",
      "carta.com",
      "cdocicmrequest.corp.microsoft.com/v-venbay@microsoft.com",
      "cgifederal.secure.force.com/bvkchaitanya@gmail.com",
      "cgifederal.secure.force.com/i_chaitu@yahoo.com",
      "cibng.ibanking-services.com/bvkchaitanya81",
      "clegc-gckey.gc.ca/bvkchaitanya",
      "cloudsimple.pagerduty.com/cbayapuneni@cloudsimple.com",
      "delivery.dhl.com/i_chaitu@yahoo.com",
      "ecams.geico.com/I_CHAITU@YAHOO.COM",
      "electroncash.slack.com/bvk",
      "eng-net-dev-01.cloudsimple.us/j.doe@devtest.cloudsimple.io",
      "eng-net-dev-03.cloudsimple.us/j.doe@devtest.cloudsimple.io",
      "env-azure-useast-devtest-08.cloudsimple.us/j.doe@devtest.cloudsimple.io",
      "env-cs-westus-devtest-04.cloudsimple.us/j.doe@devtest.cloudsimple.io",
      "env-cs-westus-devtest-08.cloudsimple.us/j.doe@devtest.cloudsimple.io",
      "env-cs-westus-devtest-51.cloudsimple.us/j.doe@devtest.cloudsimple.io",
      "exchange.bitcoin.com/bvk.other@gmail.com",
      "github.com/bvk",
      "gitlab.corp.cloudsimple.com/cbayapuneni@cloudsimple.com",
      "healthy.kaiserpermanente.org/bvkchaitanya",
      "help.steampowered.com",
      "hitbtc.com/i_chaitu@yahoo.com",
      "honest.cash/i_chaitu@yahoo.com",
      "id.atlassian.com/chaitanya@cloudsimple.com",
      "idaas-cdn-prd.balglobal.com/cbayapuneni@cloudsimple.com",
      "identity.pennymacusa.com/bvkchaitanya",
      "indiacashandcarry.com/bvkchaitanya@gmail.com",
      "link.intuit.com/bvkchaitanya@gmail.com",
      "login.comcast.net/bvkchaitanya",
      "login.live.com/chaitanya@cloudsimple.com",
      "login.microsoftonline.com/cbayapuneni@cloudsimple.com",
      "login.xfinity.com/bvkchaitanya",
      "login.yahoo.com/i_chaitu",
      "memo.cash/bvk",
      "mint.intuit.com/bvk.chaitanya",
      "msft.sts.microsoft.com/b-ilbeye@microsoft.com",
      "msft.sts.microsoft.com/v-venbay@microsoft.com",
      "msstaff.course4you.com/cbayapuneni@cloudsimple.com",
      "msstaff.course4you.com/i_chaitu@yahoo.com",
      "myaccount.google.com/cbayapuneni@cloudsimple.com",
      "myaccount.sulekha.com/I_CHAITU@YAHOO.COM",
      "myaccounts.hsabank.com/vbayapuneni2785",
      "mydoctor.kaiserpermanente.org/bvkchaitanya",
      "myinsurancegei.stillwaterinsurance.com/bvkchaitanya",
      "news.ycombinator.com/iambvk",
      "ops-eastus-csos.portal.cloudsimple.com/admin",
      "ops-eastus-stg-csos.staging.cloudsimple.com/admin",
      "portal.cloudsimple.com/csadmin@ops.cloudsimple.io",
      "portal.incometaxindiaefiling.gov.in/AHKPC5882B",
      "pushover.net/bvk.other@gmail.com",
      "rdp.az.cloudsimple.io/cbayapuneni@az.cloudsimple.io",
      "reviews.bitcoinabc.org/bvk",
      "secure.meetup.com/I_CHAITU@YAHOO.COM",
      "seq-ep.prismhr.com/bvkchaitanya",
      "signin.ebay.com/I_CHAITU@YAHOO.COM",
      "staging.cloudsimple.com/csadmin@ops.cloudsimple.io",
      "thedonald.win/freesid",
      "users.grcelearning.com",
      "venmo.com/bvk.chaitanya@gmail.com",
      "visa.mfa.gov.ua/i_chaitu@yahoo.com",
      "www.alltrails.com/bvkother@gmail.com",
      "www.amazon.com/bvkchaitanya@gmail.com",
      "www.ancestry.com/i_chaitu@yahoo.com",
      "www.bitstamp.net/ymmv0338",
      "www.coinbase.com/i_chaitu@yahoo.com",
      "www.coinex.com/6692209142",
      "www.costco.com/i_chaitu@yahoo.com",
      "www.dell.com/i_chaitu@yahoo.com",
      "www.friendlyarm.com/i_chaitu@yahoo.com",
      "www.ieltsusatest.org/i_chaitu@yahoo.com",
      "www.ieltsusatest.org/ven24080576",
      "www.instacart.com/i_chaitu@yahoo.com",
      "www.kickstarter.com/VENKATA K CHAITANYA",
      "www.linkedin.com/i_chaitu@yahoo.com",
      "www.macys.com/i_chaitu@yahoo.com",
      "www.netflix.com/deepak.vankadaru@gmail.com",
      "www.overstock.com/I_CHAITU@YAHOO.COM",
      "www.pge.com/i_chaitu@yahoo.com",
      "www.reddit.com/freesid",
      "www.safeway.com/i_chaitu@yahoo.com",
      "www.splitwise.com/bvkchaitanya@gmail.com",
      "www.spotify.com",
      "www.target.com/i_chaitu@yahoo.com",
      "www.wayfair.com/i_chaitu@yahoo.com",
      "www.yours.org/bvk.other@gmail.com",
      "www1.incometaxindiaefiling.gov.in/24/08/1981",
      "www1.incometaxindiaefiling.gov.in/AHKPC5882B",
      "yobit.net/bvkother@gmail.com"
    ]
  },
  "view_file": {
    "filename": "applications.wes.org/bvk.other@gmail.com",
    "sitename": "applications.wes.org",
    "username": "bvk.other@gmail.com",
    "password": "*",
    "data": "url:https://applications.wes.org/createaccount/login/login\n"
  },
  "delete_file": null
}
`
