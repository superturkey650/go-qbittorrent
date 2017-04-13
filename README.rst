==================
go-qbittorrent
==================

Golang wrapper for qBittorrent Web API (for versions above v3.1.x).

This wrapper is based on the methods described in `qBittorrent's Official Web API Documentation <https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-Documentation>`__

This project is based on the Python wrapper by `v1k45 <https://github.com/v1k45/python-qBittorrent>`__

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
    // not required when 'Bypass from localhost' setting is active.

    torrents = qb.Torrents()

    for torrent := range torrents{
        fmt.Println(torrent.name)
    }

API methods
===========

Getting torrents
----------------

- Get all ``active`` torrents::
.. code-block:: go

    qb.Torrents()

- Filter torrents::
.. code-block:: go

    filters := map[string]string{
        "filter": "downloading",
        "category": "my category",
    }
    qb.Torrents(filters)
    // This will return all torrents which are currently
    // downloading and are labeled as `my category`.

    filters := map[string]string{
        "filter": paused,
        "sort": ratio,
    }
    qb.Torrents(filters)
    // This will return all paused torrents sorted by their Leech:Seed ratio.

Refer to qBittorents WEB API documentation for all possible filters.

Downloading torrents
--------------------

- Download torrents by link::
.. code-block:: go

    options := map[string]string{}
    magnetLink = "magnet:?xt=urn:btih:e334ab9ddd91c10938a7....."
    qb.DownloadFromLink(magnetLink, options)

    // Will return response object with `200:OK` status code
    // regardless of sucess of failure.

- Download multipe torrents by looping over links::
.. code-block:: go

    options := map[string]string{}
    links := [...]string{link1, link2, link3}
    for l := range links{
        qb.DownloadFromLink(l, options)
    }

- Downloading torrents by file::
.. code-block:: go

    options := map[string]string{}
    file = "path/to/file.torrent"
    qb.DownloadFromFile(file, options)

- Downloading multiple torrents by using files::
.. code-block:: go

    options := map[string]string{}
    file = [...]string{path/to/file1, path/to/file2, path/to/file3}
    qb.DownloadFromFile(file, options)

- Specifing save path for downloads::
.. code-block:: go

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
    qb.DownloadFromLink(magnetLink, options)

- Applying labels to downloads::
.. code-block:: go

    label = "secret-files ;)"
    options := map[string]string{
        "label": label
    }
    file = "path/to/file.torrent"
    qb.DownloadFromFile(file, options)

    // same for links.
    category = "anime"
    options := map[string]string{
        "label": label
    }
    magnetLink = "magnet:?xt=urn:btih:e334ab9ddd91c10938a7....."
    qb.DownloadFromLink(magnetLink, options)

Pause / Resume torrents
-----------------------

- Pausing/ Resuming all torrents::
.. code-block:: go

    qb.PauseAll()
    qb.ResumeAll()

- Pausing/ Resuming a specific torrent::
.. code-block:: go

    infoHash = "e334ab9ddd....infohash....5d7fff526cb4"
    qb.Pause(infoHash)
    qb.Resume(infoHash)

- Pausing/ Resuming multiple torrents::
.. code-block:: go

    infoHashes = [...]string{
        "e334ab9ddd9......infohash......fff526cb4},
        "c9dc36f46d9......infohash......90ebebc46",
        "4c859243615......infohash......8b1f20108",
    }

    qb.PauseMultiple(infoHashes)
    qb.ResumeMultiple(infoHashes)


Full API method documentation
=============================

All API methods of qBittorrent are mentioned in docs.txt

Authors
=======

Maintainer
----------

- `Jared Mosley (jaredlmosley) <https://www.github.com/jaredlmosley/>`__

Contributors
------------

- Your name here :)

TODO
====

- Write tests
