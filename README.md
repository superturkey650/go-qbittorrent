==================
go-qBittorrent
==================

Golang wrapper for qBittorrent Web API (for versions above v3.1.x).

This wrapper is based on the methods described in `qBittorrent's Official Web API Documentation <https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-Documentation>`__

Some methods are only supported in qBittorent's latest version (v3.3.1 when writing).

It'll be best if you upgrade your client to a latest version.

Installation
============

The best way is to install with go get::

    $ go get github.com/jaredlmosley/go-qbittorrent/qbit


Quick usage guide
=================
.. code-block:: go

    import (
        "jaredlmosely/go-qbittorrent/qbit
    )

    qb := qbit.NewClient("http://localhost:8080/")

	qb.Login("admin", "your-secret-password")
    # not required when 'Bypass from localhost' setting is active.
    # defaults to admin:admin.

    torrents = qb.Torrents()

    for torrent := range torrents{
        fmt.Println(torrent.name)
    }

API methods
===========

Getting torrents
----------------

- Get all ``active`` torrents::

    qb.Torrents()

- Filter torrents::

    filters := make(map[string]string){
        "filter": "downloading",
        "category": "my category",
    }
    qb.Torrents(filters)
    // This will return all torrents which are currently
    // downloading and are labeled as ``my category``.

    filters := make(map[string]string){
        "filter": paused,
        "sort": ratio,
    }
    qb.Torrents(filters)
    // This will return all paused torrents sorted by their Leech:Seed ratio.

Refer qBittorents WEB API documentation for all possible filters.

Downloading torrents
--------------------

- Download torrents by link::

    options := map[string]string{}
    magnetLink = "magnet:?xt=urn:btih:e334ab9ddd91c10938a7....."
    qb.DownloadFromLink(magnetLink, options)

    # No matter the link is correct or not,
    # method will always return empty JSON object.

- Download multipe torrents by looping over links::

    options := map[string]string{}
    links := [...]string{link1, link2, link3}
    for l := range links{
        qb.FownloadFromLink(l, options)
    }

- Downloading torrents by file::

    options := map[string]string{}
    file = "path/to/file.torrent"
    qb.DownloadFromFile(file, options)

- Downloading multiple torrents by using files::

    options := map[string]string{}
    file = [...]string{path/to/file1, path/to/file2, path/to/file3}
    qb.DownloadFromFile(file, options)

- Specifing save path for downloads::

    savePath = "/home/user/Downloads/special-dir/"
    options := map[string]string{
        "savepath": savePath
    }
    file = "path/to/file.torrent"
    qb.DownloadFromFile(file, options)

    // same for links.
    savePath = "/home/user/Downloads/special-dir/"
    options := map[string]string{
        "savepath": savePath
    }
    magnetLink = "magnet:?xt=urn:btih:e334ab9ddd91c10938a7....."
    qb.downloadFromLink(magnetLink, options)

- Applying labels to downloads::

    label = "secret-files ;)"
    options := map[string]string{
        "label": label
    }
    file = "path/to/file.torrent"
    qb.downloadFromFile(file, options)

    // same for links.
    category = "anime"
    options := map[string]string{
        "label": label
    }
    magnetLink = "magnet:?xt=urn:btih:e334ab9ddd91c10938a7....."
    qb.downloadFromLink(magnetLink, options)

Pause / Resume torrents
-----------------------

- Pausing/ Resuming all torrents::

    qb.pause_all()
    qb.resume_all()

- Pausing/ Resuming a speicific torrent::

    info_hash = 'e334ab9ddd....infohash....5d7fff526cb4'
    qb.pause(info_hash)
    qb.resume(info_hash)

- Pausing/ Resuming multiple torrents::

    info_hash_list = ['e334ab9ddd9......infohash......fff526cb4',
                      'c9dc36f46d9......infohash......90ebebc46',
                      '4c859243615......infohash......8b1f20108']

    qb.pause_multiple(info_hash_list)
    qb.resume_multipe(info_hash_list)


Full API method documentation
=============================

All API methods of qBittorrent are mentioned @ `Read the docs <http://python-qbittorrent.readthedocs.org/en/latest/?badge=latest>`__

Authors
=======

Maintainer
----------

- `Vikas Yadav (v1k45) <https://www.github.com/v1k45/>`__

Contributors
------------

*By chronological order*

- `Matt Smith (psykzz) <https://github.com/psykzz>`__
- `Nicolas Wright (dozedoff) <https://github.com/dozedoff>`__
- `sbivol <https://github.com/sbivol>`__
- Your name here :)

TODO
====

- Write tests
