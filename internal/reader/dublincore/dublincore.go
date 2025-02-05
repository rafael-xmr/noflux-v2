// SPDX-FileCopyrightText: Copyright The Noflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package dublincore // import "github.com/fiatjaf/noflux/internal/reader/dublincore"

type DublinCoreChannelElement struct {
	DublinCoreCreator string `xml:"http://purl.org/dc/elements/1.1/ creator"`
}

type DublinCoreItemElement struct {
	DublinCoreTitle   string `xml:"http://purl.org/dc/elements/1.1/ title"`
	DublinCoreDate    string `xml:"http://purl.org/dc/elements/1.1/ date"`
	DublinCoreCreator string `xml:"http://purl.org/dc/elements/1.1/ creator"`
	DublinCoreContent string `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
}
