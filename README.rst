==================
go-qbittorrent
==================

Golang wrapper for qBittorrent Web API (for versions above v5.0.x).

This wrapper is based on the methods described in `qBittorrent's Official Web API Documentation <https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-Documentation>`__

This project is based on the Python wrapper by `v1k45 <https://github.com/v1k45/python-qBittorrent>`__

Some methods are only supported in qBittorent's latest version (v5.0.5 when writing).

It'll be best if you upgrade your client to the latest version.

An example can be found in main.go

Installation
============

The best way is to install with go get::

    $ go get github.com/superturkey650/go-qbittorrent/qbt


Quick usage guide
=================
.. code-block:: go

    import (
        "superturkey650/go-qbittorrent/qbt
    )

    qb := qbt.NewClient("http://localhost:8080/")

	qb.Login("admin", "your-secret-password")
    // not required when 'Bypass from localhost' setting is active.

    torrents = qb.Torrents()

    for torrent := range torrents{
        fmt.Println(torrent.Name)
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

    filters := qbt.TorrentsOptions{
		Filter: ptrString("downloading"),
        Category: ptrString("my category"),
	}
    qb.Torrents(filters)
    // This will return all torrents which are currently
    // downloading and are labeled as `my category`.

    filters := qbt.TorrentsOptions{
		Filter: ptrString("paused"),
        Sort: ptrString("ratio"),
	}
    qb.Torrents(filters)
    // This will return all paused torrents sorted by their Leech:Seed ratio.

Refer to qBittorents WEB API documentation for all possible filters.

Downloading torrents
--------------------

- Download torrents by link::
.. code-block:: go

    options := qbt.DownloadOptions{}
    magnetLinks = []string{"magnet:?xt=urn:btih:e334ab9ddd91c10938a7....."}
    qb.DownloadLinks(magnetLinks, options)

- Downloading torrents by file::
.. code-block:: go

    options := qbt.DownloadOptions{}
    files = []string{"path/to/file.torrent"}
    qb.DownloadFiles(files, options)

- Downloading multiple torrents by using files::
.. code-block:: go

    options := qbt.DownloadOptions{}
    files = []string{"path/to/file1", "path/to/file2", "path/to/file3"}
    qb.DownloadFiles(files, options)

- Specifing save path for downloads::
.. code-block:: go

    savePath = "/home/user/Downloads/special-dir/"
    options := qbt.DownloadOptions{
        Savepath: ptrString(savePath),
    }
    files = []string{"path/to/file.torrent"}
    qb.DownloadFiles(files, options)

    // same for links.
    savePath = "/home/user/Downloads/special-dir/"
    options := qbt.DownloadOptions{
        Savepath: ptrString(savePath),
    }
    magnetLinks = []string{"magnet:?xt=urn:btih:e334ab9ddd91c10938a7....."}
    qb.DownloadLinks(magnetLinks, options)

- Applying labels to downloads::
.. code-block:: go

    label = "secret-files"
    options := qbt.DownloadOptions{
        Label: ptrString(label),
    }
    files = []string{"path/to/file.torrent"}
    qb.DownloadFiles(files, options)

    // same for links.
    category = "anime"
    options := qbt.DownloadOptions{
        Category: ptrString(category),
    }
    magnetLinks = []string{"magnet:?xt=urn:btih:e334ab9ddd91c10938a7....."}
    qb.DownloadLinks(magnetLinks, options)

Pause / Resume torrents
-----------------------

- Pausing/ Resuming multiple torrents::
.. code-block:: go

    infoHashes = [...]string{
        "e334ab9ddd9......infohash......fff526cb4",
        "c9dc36f46d9......infohash......90ebebc46",
        "4c859243615......infohash......8b1f20108",
    }

    qb.Pause(infoHashes)
    qb.Resume(infoHashes)


Maintainer
----------

- `Jared Mosley (jaredlmosley) <https://www.github.com/superturkey650/>`__

Contributors
------------

- Your name here :)

TODO
====

- Write tests
- Implement RSS Endpoints
- Implement Search Endpoints