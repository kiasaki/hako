PV = {};
PV.Views = {};
PV.Actions = {};
PV.state = {
  encryptedContents: null,
  contents: null,
  unlocked: false,
  unlockError: false,
  inItemCreate: false,
  addItem: null,
  selectedItem: null,
  editingItem: null,
  search: "",
};

// compact uuid4 from https://gist.github.com/jed/982883
PV.uuid4 = (crypto && crypto.getRandomValues)
  ? function b(a){return a?(a^crypto.getRandomValues(new Uint8Array(1))[0]%16>>a/4).toString(16):([1e7]+-1e3+-4e3+-8e3+-1e11).replace(/[018]/g,b)}
  : function b(a){return a?(a^Math.random()*16>>a/4).toString(16):([1e7]+-1e3+-4e3+-8e3+-1e11).replace(/[018]/g,b)};

var EMPTY_VAULT = function() { return {
  v: 1,
  items: [],
}; };

var EMPTY_ITEM = function() { return {
    id: PV.uuid4(),
    name: "",
    username: "",
    password: "",
    notes: "",
} };

function initPasswordVault(filePath, contents) {
  PV.state.filePath = filePath;
  PV.state.encryptedContents = contents || null;
  PV.state.contents = null;

  var root = window.vaultRoot;
  m.mount(root, PV.Views.Top);
}

PV.Actions.saveVault = function() {
  var encryptedContents = sjcl.encrypt(PV.state.password, JSON.stringify(PV.state.contents));
  var form = new FormData();
  form.append("contents", encryptedContents);
  return m.request({
      method: "POST",
      url: "/e/" + PV.state.filePath,
      data: form,
      headers: {"Accept": "application/json"},
      withCredentials: true,
  });
};

PV.Actions.create = function(e) {
  e.preventDefault();
  PV.state.contents = EMPTY_VAULT();
  return PV.Actions.saveVault().then(function() {
    PV.state.unlocked = true;
  });
};

PV.Actions.unlock = function(e) {
  e.preventDefault();
  try {
    PV.state.contents = JSON.parse(sjcl.decrypt(PV.state.password, PV.state.encryptedContents));
  } catch (e) {
    PV.state.unlockError = true;
    return;
  }
  PV.state.unlocked = true;
  PV.state.unlockError = false;
};

PV.Actions.selectItem = function(id) {
  // TODO Centralize current item state
  PV.state.selectedItem = id;
  PV.state.inItemCreate = false;
  PV.state.editingItem = null;
};

PV.Actions.addItemStart = function() {
  // TODO Centralize current item state
  PV.state.selectedItem = null;
  PV.state.inItemCreate = true;
  PV.state.editingItem = null;
};

PV.Actions.addItem = function(e) {
  e.preventDefault();
  PV.state.contents.items.push(PV.state.addItem);
  return PV.Actions.saveVault().then(function() {
    // TODO Centralize current item state
    PV.state.selectedItem = PV.state.addItem.id;
    PV.state.inItemCreate = false;
    PV.state.editingItem = null;
  });
};

PV.Actions.editItemStart = function(id) {
  var item = R.find(
    R.propEq('id', id),
    PV.state.contents.items
  );
  // TODO Centralize current item state
  PV.state.selectedItem = null;
  PV.state.inItemCreate = false;
  PV.state.editingItem = R.clone(item);
};

PV.Actions.editItem = function(e) {
  e.preventDefault();

  var item = PV.state.editingItem;
  var itemFromStore = R.find(
    R.propEq('id', item.id),
    PV.state.contents.items
  );
  if (!itemFromStore) throw new Error("Trying to edit a non-existing item");

  itemFromStore.name = item.name;
  itemFromStore.username = item.username;
  itemFromStore.password = item.password;
  itemFromStore.notes = item.notes;

  return PV.Actions.saveVault().then(function() {
    PV.state.editingItem = null;
    PV.state.selectedItem = item.id;
  });
};

PV.Actions.deleteItem = function(id) {
  if (confirm("Are you sure?")) {
    PV.state.selectedItem = null;
    PV.state.contents.items = R.reject(
      R.propEq("id", id),
      PV.state.contents.items
    );
    return PV.Actions.saveVault();
  }
};

PV.Views.Top = {
view: function() {
  if (!PV.state.encryptedContents && !PV.state.contents) {
    return m(PV.Views.Create);
  } else if (!PV.state.unlocked) {
    return m(PV.Views.Unlock);
  } else {
    return m(PV.Views.Main);
  }
},
};

PV.Views.Button = {
view: function(vnode) {
  return m("button", {
    type: "submit",
    class: "f6 link dim br1 ph3 pv2 mb2 dib white bg-black bw0",
    onclick: vnode.attrs.onclick,
  }, vnode.attrs.text)
},
}

PV.Views.Create = {
  view: function() {
    return m("main.pa3", [
      m("p", "Chose a password for this new password vault."),
      m("p", "Passwords are never transmited to the server, if you loose it you can't recover it."),
      m("form", {onsubmit: PV.Actions.create}, [
        m("input", {
          type: "password",
          placeholder: "Vault password",
          class: "db w5 pa2 mb2 br1 ba",
          oninput: m.withAttr("value", function(value) {PV.state.password = value;}),
        }),
        m(PV.Views.Button, {text: "Create"}),
      ]),
    ]);
  }
};

PV.Views.Unlock = {
  oncreate: function() {
    document.getElementById("unlockPassword").focus();
  },
  view: function() {
    return m("form.pa3", {onsubmit: PV.Actions.unlock}, [
      m("h2.fw3.mt0", "Unlock vault"),
      PV.state.unlockError ? m("p.mt0.dark-red", "Wrong password.") : null,
      m("input", {
        id: "unlockPassword",
        type: "password",
        placeholder: "Vault password",
        class: "db w5 pa2 mb2 br1 ba",
        oninput: m.withAttr("value", function(value) {PV.state.password = value;}),
      }),
      m(PV.Views.Button, {text: "Unlock"}),
    ]);
  }
};

PV.Views.Main = {
  view: function() {
    var itemView = m(PV.Views.ItemEmptyState);
    if (PV.state.inItemCreate) {
      itemView = m(PV.Views.ItemCreate);
    } else if (PV.state.editingItem) {
      itemView = m(PV.Views.ItemEdit);
    } else if (PV.state.editingItem) {
    } else if (PV.state.selectedItem) {
      itemView = m(PV.Views.ItemView);
    }
    return m("main.cf", [
      m("aside.fl-ns.w-30-ns.br-ns.b--moon-gray", m(PV.Views.ItemList)),
      m("article.fl-ns.w-70-ns", itemView),
    ]);
  }
};

PV.Views.ItemList = {
  oncreate: function() {
    document.getElementById("searchText").focus();
  },
  selectFirst: function(filteredItems) {
    return function(e) {
      e.preventDefault();
      if (filteredItems.length > 0) {
        PV.state.selectedItem = filteredItems[0].id;
      }
    };
  },
  view: function() {
    var filteredItems = R.sortBy(R.prop("name"), R.filter(
      R.compose(R.contains(PV.state.search.toLowerCase()), R.compose(R.toLower, R.prop("name"))),
      PV.state.contents.items
    ));
    return m("div", [
      m(".pa2.tc.bb.b--moon-gray", [
        m("a", {onclick: PV.Actions.addItemStart}, "+ New item"),
      ]),
      m(".pa2.bg-near-white.bb.b--moon-gray", [
        m("form", {onsubmit: PV.Views.ItemList.selectFirst(filteredItems)}, [
          m("input.db.pa2.br1.ba.b--moon-gray.w-100", {
            id: "searchText",
            type: "text",
            placeholder: "Search...",
            oninput: m.withAttr("value", function(value) {PV.state.search = value;}),
            value: PV.state.search,
          }),
        ]),
      ]),
      m("div.overflow-y-auto", filteredItems.map(function(item) {
        return m("div.pa2.bb.b--moon-gray", {
          style: "cursor: pointer;",
          class: item.id === PV.state.selectedItem ? "light-purple bg-near-white" : "",
          onclick: PV.Actions.selectItem.bind(null, item.id),
        }, item.name);
      })),
    ]);
  },
};

PV.Views.ItemEmptyState = {
  view: function() {
    return m(".pa3", [
      m(".pb4.pt1.br2.bg-near-white.tc", [
        m("h2.fw3", "No item selected."),
        m(".i", "Select an item from the list or create one.")
      ])
    ]);
  },
};

PV.Views.ItemCreate = {
  oninit: function() {
    PV.state.addItem = EMPTY_ITEM();
  },
  view: function() {
    return m("main.pa3", [
      m(PV.Views.ItemForm, {
        item: PV.state.addItem,
        onsubmit: PV.Actions.addItem,
        onupdate: function(key, value) {
          PV.state.addItem[key] = value;
        },
      }),
    ]);
  },
};

PV.Views.ItemEdit = {
  view: function() {
    return m("main.pa3", [
      m(PV.Views.ItemForm, {
        item: PV.state.editingItem,
        onsubmit: PV.Actions.editItem,
        onupdate: function(key, value) {
          PV.state.editingItem[key] = value;
        },
      }),
    ]);
  },
};

PV.Views.ItemView = {
  oninit: function(vnode) {
    this.showPassword = false;
    this.onShow = function() {
      this.showPassword = true;
      m.redraw();
      vnode.dom.querySelector("#password").select();
    }.bind(this);
    this.onHide = function() {
      this.showPassword = false;
    }.bind(this);
  },
  onupdate: function() {
    if (PV.state.selectedItem != this.lastSelectedItem) {
      this.showPassword = false;
      m.redraw();
    }
    this.lastSelectedItem = PV.state.selectedItem;
  },
  view: function() {
    var selectedItem = R.find(
      R.propEq('id', PV.state.selectedItem),
      PV.state.contents.items
    );

    var password = ["************ ", m("a", {onclick: this.onShow}, "Show")];
    if (this.showPassword) {
      password = [
        m("input.bn.bg-near-white#password", {
          type: "text",
          readonly: true,
          onfocus: function() { this.select(); },
          onclick: function() { this.select(); },
          style: {
            width: (selectedItem.password.length * 7.5) + 'px',
            maxWidth: '400px',
          },
          value: selectedItem.password,
        }),
        " ",
        m("a", {onclick: this.onHide}, "Hide")
      ];
    }
    return m("main.pa3", [
      m("strong.db.mb1", "Name"),
      m("div.code.pa2.mb3.br1.bg-near-white", selectedItem.name || m.trust("&nbsp;")),
      m("strong.db.mb1", "Username"),
      m("div.code.pa2.mb3.br1.bg-near-white", selectedItem.username || m.trust("&nbsp;")),
      m("strong.db.mb1", "Password"),
      m("div.code.pa2.mb3.br1.bg-near-white", password),
      m("strong.db.mb1", "Notes"),
      m("pre.code.pa2.mb3.br1.bg-near-white", selectedItem.notes || m.trust("&nbsp;")),
      m(PV.Views.Button, {
        text: "Edit",
        onclick: PV.Actions.editItemStart.bind(null, selectedItem.id),
      }),
      m("a.ml2", {
        onclick: PV.Actions.deleteItem.bind(null, selectedItem.id),
      }, "Delete"),
    ]);
  },
};

PV.Views.ItemForm = {
  view: function(vnode) {
    return m("form", {onsubmit: vnode.attrs.onsubmit}, [
      m("label.pb1", "Name"),
      m("input", {
        type: "text",
        placeholder: "e.g. Email",
        class: "db w5 pa2 mb2 br1 ba",
        oninput: m.withAttr("value", vnode.attrs.onupdate.bind(null, "name")),
        value: vnode.attrs.item.name,
      }),
      m("label.pb1", "Username"),
      m("input", {
        type: "text",
        placeholder: "e.g. john.doe@mail.com",
        class: "db w5 pa2 mb2 br1 ba",
        oninput: m.withAttr("value", vnode.attrs.onupdate.bind(null, "username")),
        value: vnode.attrs.item.username,
      }),
      m("label.pb1", "Password"),
      m("input", {
        type: "text",
        placeholder: "e.g. ••••••••••••••••",
        class: "db w5 pa2 mb2 br1 ba",
        oninput: m.withAttr("value", vnode.attrs.onupdate.bind(null, "password")),
        value: vnode.attrs.item.password,
      }),
      m("label.pb1", "Notes"),
      m("textarea", {
        rows: 4,
        class: "db w-100 pa2 mb2 br1 ba",
        oninput: m.withAttr("value", vnode.attrs.onupdate.bind(null, "notes")),
        value: vnode.attrs.item.notes,
      }),
      m(PV.Views.Button, {text: "Save"}),
    ]);
  },
};
