select d.word as word, l.cmean as cmean, l.amean as amean, d.input as input, g.score as score from dictation1 d join 'grade1.5' g on d.word=g.word join list7 l on g.word=l.word order by score;
